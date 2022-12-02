package util

import (
	pb "github.com/d-ion/isglb/proto"
)

type ClientNeededSessions []*pb.ClientNeededSession

func (clients ClientNeededSessions) ToDisorderSetItemList() DisorderSetItemList[*pb.ClientNeededSession] {
	list := make([]*pb.ClientNeededSession, len(clients))
	for i, client := range clients {
		list[i] = client
	}
	return list
}

type ForwardTracks []*pb.ForwardTrack

func (tracks ForwardTracks) ToDisorderSetItemList() DisorderSetItemList[*pb.ForwardTrack] {
	list := make([]*pb.ForwardTrack, len(tracks))
	for i, track := range tracks {
		list[i] = track
	}
	return list
}

type ProceedTracks []*pb.ProceedTrack

func (tracks ProceedTracks) ToDisorderSetItemList() DisorderSetItemList[*pb.ProceedTrack] {
	indexDataList := make([]*pb.ProceedTrack, len(tracks))
	for i, track := range tracks {
		indexDataList[i] = track
	}
	return indexDataList
}

type ClientSessionItemList DisorderSetItemList[*pb.ClientNeededSession]

func (list ClientSessionItemList) ToClientSessions() []*pb.ClientNeededSession {
	tracks := make([]*pb.ClientNeededSession, len(list))
	for i, data := range list {
		tracks[i] = data
	}
	return tracks
}

type ForwardTrackItemList DisorderSetItemList[*pb.ForwardTrack]

func (list ForwardTrackItemList) ToForwardTracks() []*pb.ForwardTrack {
	tracks := make([]*pb.ForwardTrack, len(list))
	for i, data := range list {
		tracks[i] = data
	}
	return tracks
}

type ProceedTrackItemList DisorderSetItemList[*pb.ProceedTrack]

func (list ProceedTrackItemList) ToProceedTracks() []*pb.ProceedTrack {
	tracks := make([]*pb.ProceedTrack, len(list))
	for i, data := range list {
		tracks[i] = data
	}
	return tracks
}
