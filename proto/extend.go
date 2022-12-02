package proto

import (
	"fmt"
	"github.com/d-ion/isglb"
	"github.com/d-ion/isglb/util"
	"google.golang.org/protobuf/proto"
)

// Key implement isglb.Status.Key
func (s *SFUStatus) Key() string {
	return s.SFU.Id
}

// Compare implement isglb.Status.Compare
func (s *SFUStatus) Compare(status isglb.Status) bool {
	return s.String() == status.String()
}

// Clone implement isglb.Status.Clone
func (s *SFUStatus) Clone() isglb.Status {
	return proto.Clone(s).(*SFUStatus)
}

// Clone implement isglb.Report.Clone
func (r *QualityReport) Clone() isglb.Report {
	return proto.Clone(r).(*QualityReport)
}

func (i *ClientNeededSession) Key() string {
	return fmt.Sprintf("{ User: %s, Session: %s }", i.User, i.Session)
}
func (i *ClientNeededSession) Compare(data util.DisorderSetItem) bool {
	return i.String() == data.(*ClientNeededSession).String()
}
func (i *ClientNeededSession) Clone() util.DisorderSetItem {
	return proto.Clone(i).(*ClientNeededSession)
}

func (i *ForwardTrack) Key() string {
	return fmt.Sprintf("{ NID: %s, RemoteSessionId: %s }", i.Src.Id, i.RemoteSessionId)
	// !!!重要!!!不允许多次转发同一个节点的同一个Session
}

func (i *ForwardTrack) Compare(data util.DisorderSetItem) bool {
	return i.String() == data.(*ForwardTrack).String()
}

func (i *ForwardTrack) Clone() util.DisorderSetItem {
	return proto.Clone(i).(*ForwardTrack)
}

func (i *ProceedTrack) Key() string {
	return i.DstSessionId
	// !!!重要!!!不允许多个处理结果放进一个Session里
}
func (i *ProceedTrack) Compare(data util.DisorderSetItem) bool {
	srcTrackList1 := util.Strings(data.(*ProceedTrack).SrcSessionIdList).ToDisorderSetItemList()
	srcTrackSet1 := util.NewDisorderSetFromList[util.StringDisorderSetItem](srcTrackList1)
	srcTrackList2 := util.Strings(i.SrcSessionIdList).ToDisorderSetItemList()
	if !srcTrackSet1.IsSame(srcTrackList2) {
		return false
	}
	return i.String() == data.(*ProceedTrack).String()
}
func (i *ProceedTrack) Clone() util.DisorderSetItem {
	return proto.Clone(i).(*ProceedTrack)
}
