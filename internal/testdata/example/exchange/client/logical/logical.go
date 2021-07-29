package logical

import (
	"context"
	"example/exchange"

	"github.com/luno/weld"
)

var _ exchange.Client = (*client)(nil)

var Provider = weld.NewSet(New, weld.Bind(new(exchange.Client), new(*client)))

type client struct {
	b Backends
}

func New(b Backends) *client {
	return &client{
		b: b,
	}
}

func (c *client) Ping(ctx context.Context) error {
	return nil
}
