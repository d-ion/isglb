package isglb

import (
	"github.com/d-ion/isglb/proto"
	"io"
	"sync"

	"github.com/yindaheng98/execlock"
	"github.com/yindaheng98/setmap"
)

type ServerConn[S proto.Status] interface {
	setmap.IfaceSetMapItem[string]
	Send(S) error
	Recv() (proto.Request, error)
}

// Service represents isglb node
type Service[S proto.Status] struct {
	Alg proto.Algorithm // The core algorithm

	recvCh   chan isglbRecvMessage[S]
	recvChMu *execlock.SingleExec

	sendChs   map[string]chan S
	sendChsMu *sync.RWMutex
}

func NewService[S proto.Status](alg proto.Algorithm) *Service[S] {
	recvChMu := make(chan bool, 1)
	recvChMu <- true
	return &Service[S]{
		Alg:       alg,
		recvCh:    make(chan isglbRecvMessage[S], 4096),
		recvChMu:  execlock.NewSingleExec(),
		sendChs:   make(map[string]chan S),
		sendChsMu: &sync.RWMutex{},
	}
}

// isglbRecvMessage represents the message flow in Service.recvCh
// the Status and a channel receive response
type isglbRecvMessage[S proto.Status] struct {
	request proto.Request
	sigkey  ServerConn[S]
	deleted ServerConn[S]
}

// Sync receive current Status, call the algorithm, and reply expected SFUStatus
func (isglb *Service[S]) Sync(sig ServerConn[S]) error {
	skey := sig
	defer func(skey ServerConn[S]) {
		// 当连接断开的时候直接删除节点
		isglb.recvCh <- isglbRecvMessage[S]{
			deleted: skey,
		}
	}(skey)
	sendCh := make(chan S)
	isglb.sendChsMu.Lock()
	isglb.sendChs[skey.ID()] = sendCh // Create send channel when begin
	isglb.sendChsMu.Unlock()
	defer func(isglb *Service[S], skey ServerConn[S]) {
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
		isglb.recvCh <- isglbRecvMessage[S]{
			request: req,
			sigkey:  sig,
		}
	}
}

// routineStatusRecv should NOT run more than once
func (isglb *Service[S]) routineStatusRecv() {
	WhereToSend := setmap.NewHalfIfaceSetMapaMteS[string, string]()
	latestStatus := make(map[string]proto.Status) // Just for filter out those unchanged Status
	for {
		var recvCount = 0
		savedReports := make(map[string]proto.Report) // Just for filter out those deprecated reports
	L:
		for {
			var msg isglbRecvMessage[S]
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
			case *proto.RequestReport:
				log.Debugf("Received a QualityReport: %s", request.Report.String())
				if _, ok = savedReports[request.Report.String()]; !ok { //filter out deprecated report
					savedReports[request.Report.String()] = request.Report.Clone() // Save the copy
					recvCount++                                                    // count the message
				}
			case *proto.RequestStatus:
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
		statuss := make([]proto.Status, len(latestStatus))
		i = 0
		for _, s := range latestStatus {
			statuss[i] = s.Clone()
			i++
		}
		i = 0
		reports := make([]proto.Report, len(savedReports))
		for _, r := range savedReports {
			reports[i] = r
			i++
		}
		expectedStatusList := isglb.Alg.UpdateStatus(statuss, reports) // update algorithm
		expectedStatusDict := make(map[string]S, len(expectedStatusList))
		for _, expectedStatus := range expectedStatusList {
			item := expectedStatus
			expectedStatusDict[item.Key()] = item.Clone().(S) // Copy the message
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

func routineStatusSend[S proto.Status](sig ServerConn[S], sendCh <-chan S) {
	latestStatusChs := make(map[string]chan S)
	defer func(latestStatusChs map[string]chan S) {
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
			latestStatusCh = make(chan S, 1)
			latestStatusChs[msg.Key()] = latestStatusCh //Then create it

			//and create the sender goroutine
			go func(latestStatusCh <-chan S) {
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
