package proto

import "github.com/d-ion/isglb/util"

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
