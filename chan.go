package isglb

import (
	"github.com/d-ion/isglb/proto"
	"github.com/d-ion/stream"
)

type chanPair[S proto.Status] struct {
	c2s chan proto.Request
	s2c chan S
}

// ChanServerConn a ServerConn based on chan for test
type ChanServerConn[S proto.Status] struct {
	id string
	cp chanPair[S]
}

func (c ChanServerConn[S]) ID() string {
	return c.id
}

func (c ChanServerConn[S]) Send(status S) error {
	c.cp.s2c <- status
	return nil
}

func (c ChanServerConn[S]) Recv() (proto.Request, error) {
	return <-c.cp.c2s, nil
}

// ChanClientStream a ClientStream based on chan for test
type ChanClientStream[S proto.Status] struct {
	cp chanPair[S]
}

func (c ChanClientStream[S]) Send(request proto.Request) error {
	c.cp.c2s <- request
	return nil
}

func (c ChanClientStream[S]) Recv() (S, error) {
	return <-c.cp.s2c, nil
}

func NewChanCSPair[S proto.Status](id string) (ServerConn[S], stream.ClientStream[proto.Request, S]) {
	cp := chanPair[S]{
		c2s: make(chan proto.Request, 16),
		s2c: make(chan S, 16),
	}
	return ChanServerConn[S]{id: id, cp: cp}, ChanClientStream[S]{cp: cp}
}
