package isglb

import (
	"context"
	"github.com/d-ion/isglb/algorithms/impl/random"
	pb "github.com/d-ion/isglb/proto"
	"github.com/d-ion/stream"
	"testing"
	"time"
)

var id = 0

type chanClientStreamFactory struct {
	service *Service[int, *pb.SFUStatus]
}

func (c chanClientStreamFactory) NewClientStream(ctx context.Context) (stream.ClientStream[Request, *pb.SFUStatus], error) {
	id++
	server, client := NewChanCSPair[int, *pb.SFUStatus](id)
	go func() {
		err := c.service.Sync(server)
		if err != nil {
			select {
			case <-ctx.Done():
			default:
				panic(err)
			}
		}
	}()
	return client, nil
}

var client = chanClientStreamFactory{service: NewService[int, *pb.SFUStatus](&random.Random{RandomTrack: true})}

func TestSyncer(t *testing.T) {
	syncer := NewSFUStatusSyncer(client, &pb.Node{Id: "test"},
		ToolBox{
			TransmissionReporter: TestTransmissionReporter{random.RandTransmissionReport{}},
			ComputationReporter:  TestComputationReporter{random.RandComputationReport{}},
			SessionTracker:       TestSessionTracker{},
		})
	syncer.Start()
	<-time.After(300 * time.Second)
	syncer.Stop()
	<-time.After(1 * time.Second)
}
