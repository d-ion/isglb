package syncer

import (
	"github.com/d-ion/isglb"
	pb "github.com/d-ion/isglb/proto"
	"sync"
	"time"
)

type StupidTrackForwarder struct {
}

func (t StupidTrackForwarder) StartForwardTrack(trackInfo *pb.ForwardTrack) {
	isglb.GetLogger().Warnf("No StartForwardTrack  in toolbox | %+v\n", trackInfo)
}

func (t StupidTrackForwarder) StopForwardTrack(trackInfo *pb.ForwardTrack) {
	isglb.GetLogger().Warnf("No StopForwardTrack    in toolbox | %+v\n", trackInfo)
}

func (t StupidTrackForwarder) ReplaceForwardTrack(oldTrackInfo *pb.ForwardTrack, newTrackInfo *pb.ForwardTrack) {
	isglb.GetLogger().Warnf("No ReplaceForwardTrack in toolbox | %+v -> %+v\n", oldTrackInfo, newTrackInfo)
}

type StupidTrackProcesser struct {
}

func (t StupidTrackProcesser) StartProceedTrack(trackInfo *pb.ProceedTrack) {
	isglb.GetLogger().Warnf("No StartProceedTrack   in toolbox | %+v\n", trackInfo)
}

func (t StupidTrackProcesser) StopProceedTrack(trackInfo *pb.ProceedTrack) {
	isglb.GetLogger().Warnf("No StopProceedTrack    in toolbox | %+v\n", trackInfo)
}

func (t StupidTrackProcesser) ReplaceProceedTrack(oldTrackInfo *pb.ProceedTrack, newTrackInfo *pb.ProceedTrack) {
	isglb.GetLogger().Warnf("No ReplaceProceedTrack in toolbox | %+v -> %+v\n", oldTrackInfo, newTrackInfo)
}

var WarnDalay = 4 * time.Second

type StupidTransmissionReporter struct {
	sync.Once
}

func (d *StupidTransmissionReporter) Bind(chan<- *pb.TransmissionReport) {
	go d.Do(func() {
		isglb.GetLogger().Warnf("No TransmissionReporter in toolbox")
		<-time.After(WarnDalay)
	})
}

type StupidComputationReporter struct {
	sync.Once
}

func (d *StupidComputationReporter) Bind(chan<- *pb.ComputationReport) {
	go d.Do(func() {
		isglb.GetLogger().Warnf("No ComputationReporter in toolbox")
		<-time.After(WarnDalay)
	})
}

type StupidSessionTracker struct {
}

func (d StupidSessionTracker) FetchSessionEvent() *SessionEvent {
	isglb.GetLogger().Warnf("No FetchSessionEvent in toolbox")
	<-time.After(WarnDalay)
	return nil
}
