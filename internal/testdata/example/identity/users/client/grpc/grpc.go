package grpc

import (
	"context"
	"example/identity/users"
	pb "example/identity/users/userspb"
	"flag"

	"google.golang.org/grpc"

	"github.com/luno/weld"
)

var addr = flag.String("users_address", "", "host:port of users gRPC service")

var _ users.Client = (*client)(nil)

var Provider = weld.NewSet(New, weld.Bind(new(users.Client), new(*client)))

type client struct {
	rpcConn   *grpc.ClientConn
	rpcClient pb.UsersClient
}

func IsEnabled() bool {
	return *addr != ""
}

func New() (*client, error) {
	var c client
	c.rpcClient = pb.NewUsersClient(nil)
	return &c, nil
}

func (c *client) Ping(ctx context.Context) error {
	_, err := c.rpcClient.Ping(ctx, &pb.Empty{})
	return err
}
