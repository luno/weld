package logical

import (
	"context"
	"example/identity/email"
	"example/identity/email/ops"

	"github.com/luno/weld"
)

var _ email.Client = (*client)(nil)

var Provider = weld.NewSet(New, weld.Bind(new(email.Client), new(*client)))

type client struct {
	b adapter
}

func New(b Backends) *client {
	return &client{
		b: adapter{Backends: b},
	}
}

func (c *client) Ping(ctx context.Context) error {
	return ops.Logic(c.b)
}
