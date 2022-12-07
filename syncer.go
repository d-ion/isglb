package isglb

import (
	pb "github.com/d-ion/isglb/proto"
	"github.com/d-ion/isglb/syncer"
	"github.com/d-ion/isglb/util"
	"google.golang.org/protobuf/proto"
)

// SFUStatusSyncer is a Client to sync SFUStatus
type SFUStatusSyncer struct {
	client  *Client[*pb.SFUStatus, *pb.QualityReport]
	descSFU *pb.Node

	router   trackRouter
	reporter *qualityReporter
	session  SessionTracker

	clientSet       *util.DisorderSet[*pb.ClientNeededSession]
	forwardTrackSet *util.DisorderSet[*pb.ForwardTrack]
	proceedTrackSet *util.DisorderSet[*pb.ProceedTrack]

	// Just recv and send latest status
	statusRecvCh   chan *pb.SFUStatus
	statusSendCh   chan bool
	sessionEventCh chan *SessionEvent
}

func NewSFUStatusSyncer(factory ClientStreamFactory[*pb.SFUStatus], myself *pb.Node, toolbox ToolBox) *SFUStatusSyncer {
	isglbClient := NewClient[*pb.SFUStatus, *pb.QualityReport](factory)
	if isglbClient == nil {
		return nil
	}
	forwarder, processor, session := toolbox.TrackForwarder, toolbox.TrackProcessor, toolbox.SessionTracker
	if forwarder == nil {
		forwarder = syncer.StupidTrackForwarder{}
	}
	if processor == nil {
		processor = syncer.StupidTrackProcesser{}
	}
	if session == nil {
		session = syncer.StupidSessionTracker{}
	}
	tr, cr := toolbox.TransmissionReporter, toolbox.ComputationReporter
	if tr == nil {
		tr = &syncer.StupidTransmissionReporter{}
	}
	if cr == nil {
		cr = &syncer.StupidComputationReporter{}
	}
	s := &SFUStatusSyncer{
		client:  isglbClient,
		descSFU: myself,

		router: trackRouter{
			TrackForwarder: forwarder,
			TrackProcessor: processor,
		},
		reporter: newQualityReporter(tr, cr),
		session:  session,

		clientSet:       util.NewDisorderSet[*pb.ClientNeededSession](),
		forwardTrackSet: util.NewDisorderSet[*pb.ForwardTrack](),
		proceedTrackSet: util.NewDisorderSet[*pb.ProceedTrack](),

		statusRecvCh:   make(chan *pb.SFUStatus, 1),
		statusSendCh:   make(chan bool, 1),
		sessionEventCh: make(chan *SessionEvent, 1024),
	}
	s.statusSendCh <- true
	isglbClient.OnStatusRecv = func(si Status) {
		st := si.(*pb.SFUStatus)
		select {
		case _, ok := <-s.statusRecvCh:
			if !ok {
				return
			}
		default:
		}
		select {
		case s.statusRecvCh <- st:
		default:
		}
	}
	return s
}

func (s *SFUStatusSyncer) NotifySFUStatus() {
	// Only send latest status
	select {
	case s.statusSendCh <- true:
	default:
	}
}

// ↓↓↓↓↓ should access Index, so keep single thread ↓↓↓↓↓

// syncStatus sync the current SFUStatus with the expected SFUStatus
// MUST be single threaded
func (s *SFUStatusSyncer) syncStatus(expectedStatus *pb.SFUStatus) {
	if expectedStatus.SFU.String() != s.descSFU.String() { // Check if the SFU status is mine
		// If not
		log.Warnf("Received SFU status is not mine, drop it: %s", expectedStatus.SFU)
		s.NotifySFUStatus() // The server must re-consider the status for our SFU
		return              // And we should wait for the right SFU status to come
	}

	// Check if the client needed Client is same
	sessionIndexDataList := pb.ClientNeededSessions(expectedStatus.Clients).ToDisorderSetItemList()
	if !s.clientSet.IsSame(sessionIndexDataList) { // Check if the ClientNeededSessions is same
		// If not
		log.Warnf("Received SFU status have different Client list, drop it: %s", expectedStatus.Clients)
		s.NotifySFUStatus() // The server must re-consider the status for our SFU
		return              // And we should wait for the right SFU status to come
	}

	// Perform track forward change
	forwardIndexDataList := pb.ForwardTracks(expectedStatus.ForwardTracks).ToDisorderSetItemList()
	forwardAdd, forwardDel, forwardReplace := s.forwardTrackSet.Update(forwardIndexDataList)
	for _, track := range forwardDel {
		s.router.StopForwardTrack(track)
	}
	for _, track := range forwardReplace {
		s.router.ReplaceForwardTrack(
			track.Old,
			track.New,
		)
	}
	for _, track := range forwardAdd {
		s.router.StartForwardTrack(track)
	}

	//Perform track proceed change
	proceedIndexDataList := pb.ProceedTracks(expectedStatus.ProceedTracks).ToDisorderSetItemList()
	proceedAdd, proceedDel, proceedReplace := s.proceedTrackSet.Update(proceedIndexDataList)
	for _, track := range proceedDel {
		s.router.StopProceedTrack(track)
	}
	for _, track := range proceedReplace {
		s.router.ReplaceProceedTrack(
			track.Old,
			track.New,
		)
	}
	for _, track := range proceedAdd {
		s.router.StartProceedTrack(track)
	}
}

// handleSessionEvent handle the SessionEvent
// MUST be single threaded
func (s *SFUStatusSyncer) handleSessionEvent(event *SessionEvent) {
	// Just add or remove it, and sand latest status
	switch event.State {
	case SessionEvent_ADD:
		s.clientSet.Add(event.Session)
		s.NotifySFUStatus()
	case SessionEvent_REMOVE:
		s.clientSet.Del(event.Session)
		s.NotifySFUStatus()
	}
}

// main is the "main function" goroutine of the NewSFUStatusSyncer
// All the methods about Index should be here, to ensure the assess is single-threaded
func (s *SFUStatusSyncer) main() {
	for {
		select {
		case event, ok := <-s.sessionEventCh: // handle an event
			if !ok {
				return
			}
			s.handleSessionEvent(event) // should access Index, so keep single thread
		case st, ok := <-s.statusRecvCh: // handle a received SFU status
			if !ok {
				return
			}
			s.syncStatus(st) // should access Index, so keep single thread
		case _, ok := <-s.statusSendCh: // handle SFU status send event
			if !ok {
				return
			}
			st := &pb.SFUStatus{
				SFU:           proto.Clone(s.descSFU).(*pb.Node),
				ForwardTracks: pb.ForwardTrackItemList(s.forwardTrackSet.Sort()).ToForwardTracks(),
				ProceedTracks: pb.ProceedTrackItemList(s.proceedTrackSet.Sort()).ToProceedTracks(),
				Clients:       pb.ClientSessionItemList(s.clientSet.Sort()).ToClientSessions(),
			} // should access Index, so keep single thread
			go s.client.SendStatus(st)
		}
	}
}

// ↑↑↑↑↑ should access Index, so keep single thread ↑↑↑↑↑

func (s *SFUStatusSyncer) sessionFetcher() {
	defer func() {
		if err := recover(); err != nil {
			log.Debugf("error on close: %+v", err)
		}
	}()
	for {
		event := s.session.FetchSessionEvent()
		if event == nil {
			return
		}
		s.sessionEventCh <- event.Clone()
	}
}

func (s *SFUStatusSyncer) reportFetcher() {
	for {
		report := s.reporter.FetchReport()
		if report == nil {
			return
		}
		go s.client.SendReport(report)
	}
}

func (s *SFUStatusSyncer) Start() {
	go s.main()
	go s.reportFetcher()
	go s.sessionFetcher()
	s.client.Connect()
}

func (s *SFUStatusSyncer) Stop() {
	s.client.Close()
	close(s.statusRecvCh)
	close(s.statusSendCh)
	close(s.sessionEventCh)
}
