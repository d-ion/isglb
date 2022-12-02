package isglb

import (
	"context"
	"github.com/d-ion/stream"
	"github.com/sirupsen/logrus"
	"github.com/yindaheng98/execlock"
)

type ClientStreamFactory interface {
	NewClientStream(ctx context.Context) (stream.ClientStream[Request, Status], error)
}

type Client struct {
	*stream.Client[Request, Status]
	ctxTop    context.Context
	cancelTop context.CancelFunc

	sendStatusExec *execlock.SingleLatestExec

	cancelLast context.CancelFunc

	OnStatusRecv func(s Status)

	Logger *logrus.Logger
}

func NewClient(factory ClientStreamFactory) *Client {
	ctx, cancal := context.WithCancel(context.Background())
	c := &Client{
		Client:         stream.NewClient[Request, Status](factory),
		ctxTop:         ctx,
		cancelTop:      cancal,
		sendStatusExec: &execlock.SingleLatestExec{},
		Logger:         logrus.StandardLogger(),
	}
	c.OnMsgRecv(func(status Status) {
		if c.OnStatusRecv != nil {
			c.OnStatusRecv(status)
		}
	})
	return c
}

// SendReport send the report, maybe lose when cannot connect
func (c *Client) SendReport(report Report) {
	c.DoWithClient(func(client stream.ClientStream[Request, Status]) error {
		err := client.Send(&RequestReport{Report: report})
		if err != nil {
			c.Logger.Errorf("Report send error: %+v", err)
			return err
		}
		return nil
	})
}

// SendStatus send the Status, if there is a new status should be send, the last send will be canceled
func (c *Client) SendStatus(status Status) {
	c.sendStatusExec.Do(func(ctx context.Context) {
		for {
			select {
			case <-c.ctxTop.Done():
				return
			case <-ctx.Done():
				return
			default:
			}
			ok := c.DoWithClient(func(client stream.ClientStream[Request, Status]) error {
				err := client.Send(&RequestStatus{Status: status})
				if err != nil {
					c.Logger.Errorf("Status send error: %+v", err)
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
