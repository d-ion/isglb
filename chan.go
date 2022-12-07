package isglb

import (
	"github.com/d-ion/stream"
)

type chanPair[S Status] struct {
	c2s chan Request
	s2c chan S
}

type ChanServerConn[T comparable, S Status] struct {
	id T
	cp chanPair[S]
}

func (c ChanServerConn[T, S]) ID() T {
	return c.id
}

func (c ChanServerConn[T, S]) Send(status S) error {
	c.cp.s2c <- status
	return nil
}

func (c ChanServerConn[T, S]) Recv() (Request, error) {
	return <-c.cp.c2s, nil
}

type ChanClientStream[S Status] struct {
	cp chanPair[S]
}

func (c ChanClientStream[S]) Send(request Request) error {
	c.cp.c2s <- request
	return nil
}

func (c ChanClientStream[S]) Recv() (S, error) {
	return <-c.cp.s2c, nil
}

func NewChanCSPair[T comparable, S Status](id T) (ServerConn[T, S], stream.ClientStream[Request, S]) {
	cp := chanPair[S]{
		c2s: make(chan Request, 16),
		s2c: make(chan S, 16),
	}
	return ChanServerConn[T, S]{id: id, cp: cp}, ChanClientStream[S]{cp: cp}
}
