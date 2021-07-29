package grpc

import (
	"context"
	"example/exchange"
	pb "example/exchange/exchangepb"
	"flag"

	"google.golang.org/grpc"

	"github.com/luno/weld"
)

var addr = flag.String("exchange_address", "", "host:port of exchange gRPC service")

var _ exchange.Client = (*client)(nil)

var Provider = weld.NewSet(New, weld.Bind(new(exchange.Client), new(*client)))

type client struct {
	rpcConn   *grpc.ClientConn
	rpcClient pb.ExchangeClient
}

func IsEnabled() bool {
	return *addr != ""
}

func New() (*client, error) {
	var c client
	c.rpcClient = pb.NewExchangeClient(nil)
	return &c, nil
}

func (c *client) Ping(ctx context.Context) error {
	_, err := c.rpcClient.Ping(ctx, &pb.Empty{})
	return err
}
