package isglb

import (
	"context"
	"fmt"
	"github.com/d-ion/isglb/algorithms/random"
	pb "github.com/d-ion/isglb/proto"
	"github.com/d-ion/stream"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"
	"time"
)

var id = 0

type chanClientStreamFactory struct {
	service *Service[*pb.SFUStatus]
}

func (c chanClientStreamFactory) NewClientStream(ctx context.Context) (stream.ClientStream[pb.Request, *pb.SFUStatus], error) {
	id++
	server, client := NewChanCSPair[*pb.SFUStatus](fmt.Sprintf("%04d", id))
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

func TestClient(t *testing.T) {
	client := NewClient[*pb.SFUStatus, *pb.QualityReport](chanClientStreamFactory{
		service: NewService[*pb.SFUStatus](
			pb.AlgorithmWrapper{ProtobufAlgorithm: &random.Random{RandomTrack: true}},
		),
	})
	client.OnStatusRecv = func(s *pb.SFUStatus) {
		t.Logf("Received %+v", s)
	}
	reporter := random.RandReports{}
	for i := 0; i < 10; i++ {
		for j := 0; j < 4; j++ {
			info, _ := anypb.New(&pb.Node{Id: random.RandomString(4)})
			client.SendStatus(&pb.SFUStatus{
				SFU: &pb.Node{
					Id:   "test-" + random.RandomString(4),
					Info: info,
				},
			})
			for _, r := range reporter.RandReports() {
				client.SendReport(r)
			}
		}
		<-time.After(2 * time.Second)
	}
	client.Close()
	<-time.After(2 * time.Second)
}
