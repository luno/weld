package grpc

import (
	"context"
	"example/identity/email"
	pb "example/identity/email/emailpb"
	"flag"

	"google.golang.org/grpc"

	"github.com/luno/weld"
)

var addr = flag.String("email_address", "", "host:port of email gRPC service")

var _ email.Client = (*client)(nil)

var Provider = weld.NewSet(New, weld.Bind(new(email.Client), new(*client)))

type client struct {
	rpcConn   *grpc.ClientConn
	rpcClient pb.EmailClient
}

func IsEnabled() bool {
	return *addr != ""
}

func New() (*client, error) {
	var c client
	c.rpcClient = pb.NewEmailClient(nil)
	return &c, nil
}

func (c *client) Ping(ctx context.Context) error {
	_, err := c.rpcClient.Ping(ctx, &pb.Empty{})
	return err
}
