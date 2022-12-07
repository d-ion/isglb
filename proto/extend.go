package proto

import (
	"google.golang.org/protobuf/proto"
)

// Key implement isglb.Status.Key
func (s *SFUStatus) Key() string {
	return s.SFU.Id
}

// Compare implement isglb.Status.Compare
func (s *SFUStatus) Compare(status Status) bool {
	return s.String() == status.String()
}

// Clone implement isglb.Status.Clone
func (s *SFUStatus) Clone() Status {
	return proto.Clone(s).(*SFUStatus)
}

// Clone implement isglb.Report.Clone
func (r *QualityReport) Clone() Report {
	return proto.Clone(r).(*QualityReport)
}
