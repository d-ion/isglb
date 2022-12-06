package isglb

import (
	"io"
	"sync"

	"github.com/yindaheng98/execlock"
	"github.com/yindaheng98/setmap"
)

type ServerConn[ServerConnIDType comparable] interface {
	setmap.IfaceSetMapItem[ServerConnIDType]
	Send(Status) error
	Recv() (Request, error)
}

// Service represents isglb node
type Service[ServerConnIDType comparable] struct {
	Alg Algorithm // The core algorithm

	recvCh   chan isglbRecvMessage[ServerConnIDType]
	recvChMu *execlock.SingleExec

	sendChs   map[ServerConnIDType]chan Status
	sendChsMu *sync.RWMutex
}

func NewService[T comparable](alg Algorithm) *Service[T] {
	recvChMu := make(chan bool, 1)
	recvChMu <- true
	return &Service[T]{
		Alg:       alg,
		recvCh:    make(chan isglbRecvMessage[T], 4096),
		recvChMu:  execlock.NewSingleExec(),
		sendChs:   make(map[T]chan Status),
		sendChsMu: &sync.RWMutex{},
	}
}

// isglbRecvMessage represents the message flow in Service.recvCh
// the Status and a channel receive response
type isglbRecvMessage[T comparable] struct {
	request Request
	sigkey  ServerConn[T]
	deleted ServerConn[T]
}

// Sync receive current Status, call the algorithm, and reply expected SFUStatus
func (isglb *Service[T]) Sync(sig ServerConn[T]) error {
	skey := sig
	defer func(skey ServerConn[T]) {
		// 当连接断开的时候直接删除节点
		isglb.recvCh <- isglbRecvMessage[T]{
			deleted: skey,
		}
	}(skey)
	sendCh := make(chan Status)
	isglb.sendChsMu.Lock()
	isglb.sendChs[skey.ID()] = sendCh // Create send channel when begin
	isglb.sendChsMu.Unlock()
	defer func(isglb *Service[T], skey ServerConn[T]) {
		isglb.sendChsMu.Lock()
		if sendCh, ok := isglb.sendChs[skey.ID()]; ok {
			close(sendCh)
			delete(isglb.sendChs, skey.ID()) // delete send channel when exit
		}
		isglb.sendChsMu.Unlock()
	}(isglb, skey)

	go routineStatusSend(sig, sendCh)          //start message sending
	isglb.recvChMu.Do(isglb.routineStatusRecv) // Do not start again

	for {
		req, err := sig.Recv() // Receive a Request
		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Errorf("SyncRequest receive error %d", err)
			return err
		}
		// Push to receive channel
		isglb.recvCh <- isglbRecvMessage[T]{
			request: req,
			sigkey:  sig,
		}
	}
}

// routineStatusRecv should NOT run more than once
func (isglb *Service[T]) routineStatusRecv() {
	WhereToSend := setmap.NewHalfIfaceSetMapaMteS[string, T]()
	latestStatus := make(map[string]Status) // Just for filter out those unchanged Status
	for {
		var recvCount = 0
		savedReports := make(map[string]Report) // Just for filter out those deprecated reports
	L:
		for {
			var msg isglbRecvMessage[T]
			var ok bool
			if recvCount <= 0 { //if there is no message
				msg, ok = <-isglb.recvCh //wait for the first message
				if !ok {                 //if closed
					return //exit
				}
			} else {
				select {
				case msg, ok = <-isglb.recvCh: // Receive a message
					if !ok { //if closed
						return //exit
					}
				default: //if there is no more message
					break L //just exit
				}
			}

			if deletedSig := msg.deleted; deletedSig != nil {
				for _, nid := range WhereToSend.GetUniqueKeys(deletedSig) {
					WhereToSend.RemoveKey(nid)
					if lastStatus, ok := latestStatus[nid]; ok {
						delete(latestStatus, nid)
						log.Debugf("Deleted a Status because its sig exit: %s", lastStatus.String())
						recvCount++ // count the message
					} else {
						log.Debugf("Status to be deleted not exists: %s", nid)
					}
				}
				WhereToSend.RemoveValue(deletedSig)
			}

			if msg.request == nil || msg.sigkey == nil {
				continue
			}
			//category and save messages
			switch request := msg.request.(type) {
			case *RequestReport:
				log.Debugf("Received a QualityReport: %s", request.Report.String())
				if _, ok = savedReports[request.Report.String()]; !ok { //filter out deprecated report
					savedReports[request.Report.String()] = request.Report.Clone() // Save the copy
					recvCount++                                                    // count the message
				}
			case *RequestStatus:
				log.Debugf("Received a Status: %s", request.Status.String())
				reportedStatus := request.Status
				nid := reportedStatus.Key()

				WhereToSend.Add(reportedStatus.Key(), msg.sigkey) // Save sig and nid

				if lastStatus, ok := latestStatus[nid]; ok && lastStatus.Compare(reportedStatus) {
					log.Debugf("Dropped deprecated SFU status from request: %s", lastStatus.String())
					continue //filter out unchanged status
				}
				// If the request has changed
				latestStatus[nid] = reportedStatus.Clone() // Save Status copy
				recvCount++                                // count the message
			}
		}

		// proceed all those received messages above
		if recvCount <= 0 { //if there is no valid message
			continue //do nothing
		}

		var i int
		statuss := make([]Status, len(latestStatus))
		i = 0
		for _, s := range latestStatus {
			statuss[i] = s.Clone()
			i++
		}
		i = 0
		reports := make([]Report, len(savedReports))
		for _, r := range savedReports {
			reports[i] = r
			i++
		}
		expectedStatusList := isglb.Alg.UpdateStatus(statuss, reports) // update algorithm
		expectedStatusDict := make(map[string]Status, len(expectedStatusList))
		for _, expectedStatus := range expectedStatusList {
			item := expectedStatus
			expectedStatusDict[item.Key()] = item.Clone() // Copy the message
		}
		for nid, expectedStatus := range expectedStatusDict {
			if lastStatus, ok := latestStatus[nid]; ok && lastStatus.Compare(expectedStatus) {
				log.Debugf("Dropped deprecated SFU status from algorithm: %s", lastStatus.String())
				continue //filter out unchanged request
			}
			// If the request should be change
			sigs := WhereToSend.GetSet(nid)
			if len(sigs) <= 0 {
				log.Warnf("No Status sender sig found for nid %s: %s", nid, expectedStatus.String())
				continue
			}
			isglb.sendChsMu.RLock()
			if sendCh, ok := isglb.sendChs[sigs[0].ID()]; ok {
				sendCh <- expectedStatus           // Send it
				latestStatus[nid] = expectedStatus // And Save it
			} else {
				log.Warnf("No Status sender channel found for nid %s: %s", nid, expectedStatus.String())
			}
			isglb.sendChsMu.RUnlock()
		}
	}
}

func routineStatusSend[T comparable](sig ServerConn[T], sendCh <-chan Status) {
	latestStatusChs := make(map[string]chan Status)
	defer func(latestStatusChs map[string]chan Status) {
		for nid, ch := range latestStatusChs {
			close(ch)
			delete(latestStatusChs, nid)
		}
	}(latestStatusChs)
	for {
		msg, ok := <-sendCh
		if !ok {
			return
		}
		latestStatusCh, ok := latestStatusChs[msg.Key()]
		if !ok { //If latest status not exists
			latestStatusCh = make(chan Status, 1)
			latestStatusChs[msg.Key()] = latestStatusCh //Then create it

			//and create the sender goroutine
			go func(latestStatusCh <-chan Status) {
				for {
					latestStatus, ok := <-latestStatusCh //get status
					if !ok {                             //if chan closed
						return //exit
					}
					// If the status should be change
					err := sig.Send(latestStatus)
					if err != nil {
						if err == io.EOF {
							return
						}
						log.Errorf("%v SFU request send error", err)
					}
				}
			}(latestStatusCh)
		}
		select {
		case latestStatusCh <- msg: //check if there is a message not send
		// no message, that's ok, our message pushed
		default: //if there is a message not send
			select {
			case <-latestStatusCh: //delete it
			default:
			}
			latestStatusCh <- msg //and push the latest message
		}
	}
}
