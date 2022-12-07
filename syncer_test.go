package isglb

import (
	"github.com/d-ion/isglb/algorithms/random"
	pb "github.com/d-ion/isglb/proto"
	"github.com/sirupsen/logrus"
	"math/rand"
	"testing"
	"time"
)

const sleep = 1000

type TestTransmissionReporter struct {
	random.RandTransmissionReport
}

func (t TestTransmissionReporter) Bind(ch chan<- *pb.TransmissionReport) {
	go func(ch chan<- *pb.TransmissionReport) {
		for {
			<-time.After(time.Duration(rand.Int31n(sleep)) * time.Millisecond)
			ch <- t.RandReport()
		}
	}(ch)
}

type TestComputationReporter struct {
	random.RandComputationReport
}

func (t TestComputationReporter) Bind(ch chan<- *pb.ComputationReport) {
	go func(ch chan<- *pb.ComputationReport) {
		for {
			<-time.After(time.Duration(rand.Int31n(sleep)) * time.Millisecond)
			ch <- t.RandReport()
		}
	}(ch)
}

type TestSessionTracker struct {
}

func (t TestSessionTracker) FetchSessionEvent() *SessionEvent {
	<-time.After(time.Duration(rand.Int31n(sleep)) * time.Millisecond)
	return &SessionEvent{
		Session: &pb.ClientNeededSession{
			Session: "",
			User:    "",
		}, State: SessionEvent_State(rand.Intn(2)),
	}
}

func TestSyncer(t *testing.T) {
	log.SetLevel(logrus.DebugLevel)
	syncer := NewSFUStatusSyncer(
		chanClientStreamFactory{
			service: NewService[*pb.SFUStatus](
				pb.AlgorithmWrapper{ProtobufAlgorithm: &random.Random{RandomTrack: true}},
			),
		},
		&pb.Node{Id: "test"},
		ToolBox{
			TransmissionReporter: TestTransmissionReporter{random.RandTransmissionReport{}},
			ComputationReporter:  TestComputationReporter{random.RandComputationReport{}},
			SessionTracker:       TestSessionTracker{},
		})
	syncer.Start()
	<-time.After(10 * time.Second)
	syncer.Stop()
	<-time.After(1 * time.Second)
}
