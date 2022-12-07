package proto

import (
	"fmt"
	"github.com/d-ion/isglb/util"
	"google.golang.org/protobuf/proto"
)

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

type ClientNeededSessions []*ClientNeededSession

func (clients ClientNeededSessions) ToDisorderSetItemList() util.DisorderSetItemList[*ClientNeededSession] {
	list := make([]*ClientNeededSession, len(clients))
	for i, client := range clients {
		list[i] = client
	}
	return list
}

type ForwardTracks []*ForwardTrack

func (tracks ForwardTracks) ToDisorderSetItemList() util.DisorderSetItemList[*ForwardTrack] {
	list := make([]*ForwardTrack, len(tracks))
	for i, track := range tracks {
		list[i] = track
	}
	return list
}

type ProceedTracks []*ProceedTrack

func (tracks ProceedTracks) ToDisorderSetItemList() util.DisorderSetItemList[*ProceedTrack] {
	indexDataList := make([]*ProceedTrack, len(tracks))
	for i, track := range tracks {
		indexDataList[i] = track
	}
	return indexDataList
}

type ClientSessionItemList util.DisorderSetItemList[*ClientNeededSession]

func (list ClientSessionItemList) ToClientSessions() []*ClientNeededSession {
	tracks := make([]*ClientNeededSession, len(list))
	for i, data := range list {
		tracks[i] = data
	}
	return tracks
}

type ForwardTrackItemList util.DisorderSetItemList[*ForwardTrack]

func (list ForwardTrackItemList) ToForwardTracks() []*ForwardTrack {
	tracks := make([]*ForwardTrack, len(list))
	for i, data := range list {
		tracks[i] = data
	}
	return tracks
}

type ProceedTrackItemList util.DisorderSetItemList[*ProceedTrack]

func (list ProceedTrackItemList) ToProceedTracks() []*ProceedTrack {
	tracks := make([]*ProceedTrack, len(list))
	for i, data := range list {
		tracks[i] = data
	}
	return tracks
}
