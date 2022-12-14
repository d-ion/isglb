package random

import (
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math/rand"

	pb "github.com/d-ion/isglb/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// Random is a node selection algorithm, just for test
type Random struct {
	nodes       map[string]*pb.SFUStatus
	RandomTrack bool
}

func RandBool() bool {
	return rand.Intn(2) == 0
}

func RandNode(nid string) *pb.Node {
	info, _ := anypb.New(&pb.Node{
		Id: nid + RandomString(4),
	})
	return &pb.Node{
		Id:   nid,
		Info: info,
	}
}

func RandChange(s *pb.SFUStatus) *pb.SFUStatus {
	info, _ := anypb.New(&pb.Node{
		Id: s.SFU.Id + RandomString(4),
	})
	s.SFU.Info = info
	return s
}

func (r *Random) UpdateSFUStatus(current []*pb.SFUStatus, reports []*pb.QualityReport) (expected []*pb.SFUStatus) {
	fmt.Printf("┎Received status: %+v\n", current)
	fmt.Printf("┖Received report: %+v\n", reports)
	if r.nodes == nil {
		r.nodes = map[string]*pb.SFUStatus{}
	}
	for _, s := range current {
		r.nodes[s.SFU.Id] = s
		if !r.RandomTrack && RandBool() { // change SFUStatus.SFU only when not random change track
			s = RandChange(s)
		}
		if RandBool() {
			expected = append(expected, s)
		}
	}

	if RandBool() {
		expected = append(expected, &pb.SFUStatus{
			SFU: RandNode("test-" + RandomString(6)),
		})
	}

	for _, s := range r.nodes {
		if !r.RandomTrack && RandBool() {
			s = RandChange(s)
		}
		if RandBool() {
			expected = append(expected, s)
		}
	}

	if r.RandomTrack {
		for _, s := range expected {
			if RandBool() {
				s.ForwardTracks = RandChangeForwardTracks(s.ForwardTracks)
			}
			if RandBool() {
				s.ProceedTracks = RandChangeProceedTracks(s.ProceedTracks)
			}
		}
	}
	fmt.Printf("▶  Return status: %+v\n", expected)
	return
}

type RandTransmissionReport struct {
	reports []*pb.TransmissionReport
}

func (r *RandTransmissionReport) RandReport() *pb.TransmissionReport {
	t, _ := anypb.New(&timestamp.Timestamp{Seconds: rand.Int63()})
	report := &pb.TransmissionReport{
		Report: t,
	}
	if RandBool() && len(r.reports) > 0 {
		report = r.reports[rand.Intn(len(r.reports))]
	} else {
		if RandBool() {
			r.reports = append(r.reports, report)
		}
	}
	if RandBool() {
		report.Report, _ = anypb.New(&timestamp.Timestamp{Seconds: rand.Int63()})
	}
	return report
}

type RandComputationReport struct {
	reports []*pb.ComputationReport
}

func (r *RandComputationReport) RandReport() *pb.ComputationReport {
	t, _ := anypb.New(&timestamp.Timestamp{Seconds: rand.Int63()})
	report := &pb.ComputationReport{
		Report: t,
	}
	if RandBool() && len(r.reports) > 0 {
		report = r.reports[rand.Intn(len(r.reports))]
	} else {
		if RandBool() {
			r.reports = append(r.reports, report)
		}
	}
	if RandBool() {
		report.Report, _ = anypb.New(&timestamp.Timestamp{Seconds: rand.Int63()})
	}
	return report
}

type RandReports struct {
	RandTransmissionReport
	RandComputationReport
}

func (r *RandReports) RandReports() []*pb.QualityReport {
	n := rand.Intn(16)
	rs := make([]*pb.QualityReport, n)
	for i := 0; i < n; i++ {
		if RandBool() {
			rs[i] = &pb.QualityReport{
				Timestamp: timestamppb.Now(),
				Report:    &pb.QualityReport_Transmission{Transmission: r.RandTransmissionReport.RandReport()},
			}
		} else {
			rs[i] = &pb.QualityReport{
				Timestamp: timestamppb.Now(),
				Report:    &pb.QualityReport_Computation{Computation: r.RandComputationReport.RandReport()},
			}
		}
	}
	return rs
}
