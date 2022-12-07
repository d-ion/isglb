package isglb

import (
	"context"
	"github.com/d-ion/isglb/proto"
	"github.com/d-ion/stream"
	"github.com/yindaheng98/execlock"
)

// ClientStreamFactory for stream.Client. Generate stream.ClientStream for reconnect
type ClientStreamFactory[S proto.Status] interface {
	NewClientStream(ctx context.Context) (stream.ClientStream[proto.Request, S], error)
}

// Client for send/recv Status/Report
type Client[S proto.Status, R proto.Report] struct {
	*stream.Client[proto.Request, S]
	ctxTop    context.Context
	cancelTop context.CancelFunc

	sendStatusExec *execlock.SingleLatestExec

	cancelLast context.CancelFunc

	OnStatusRecv func(s S)
}

func NewClient[S proto.Status, R proto.Report](factory ClientStreamFactory[S]) *Client[S, R] {
	ctx, cancal := context.WithCancel(context.Background())
	c := &Client[S, R]{
		Client:         stream.NewClient[proto.Request, S](factory),
		ctxTop:         ctx,
		cancelTop:      cancal,
		sendStatusExec: &execlock.SingleLatestExec{},
	}
	c.OnMsgRecv(func(status S) {
		if c.OnStatusRecv != nil {
			c.OnStatusRecv(status)
		}
	})
	return c
}

// SendReport send the report, maybe lose when cannot connect
func (c *Client[S, R]) SendReport(report R) {
	c.DoWithClient(func(client stream.ClientStream[proto.Request, S]) error { // Just run it
		err := client.Send(&proto.RequestReport{Report: report})
		if err != nil {
			log.Errorf("Report send error: %+v", err)
			return err
		}
		return nil
	})
}

// SendStatus send the Status, if there is a new status should be sent, the last send will be canceled
func (c *Client[S, R]) SendStatus(status S) {
	c.sendStatusExec.Do(func(ctx context.Context) { // Run until success
		for {
			select {
			case <-c.ctxTop.Done():
				return
			case <-ctx.Done():
				return
			default:
			}
			ok := c.DoWithClient(func(client stream.ClientStream[proto.Request, S]) error {
				err := client.Send(&proto.RequestStatus{Status: status})
				if err != nil {
					log.Errorf("Status send error: %+v", err)
					return err
				}
				return nil
			})
			if ok {
				return
			}
		}
	})
}
