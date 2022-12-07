package isglb

import (
	"github.com/d-ion/isglb/algorithms/random"
	pb "github.com/d-ion/isglb/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	client := NewClient[*pb.SFUStatus, *pb.QualityReport](NewChanClientStreamFactory[*pb.SFUStatus](
		NewService[*pb.SFUStatus](
			pb.AlgorithmWrapper{ProtobufAlgorithm: &random.Random{RandomTrack: true}},
		),
	))
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
