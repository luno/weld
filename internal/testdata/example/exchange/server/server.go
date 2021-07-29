package server

import (
	"context"
	pb "example/exchange/exchangepb"
)

var _ pb.ExchangeServer = (*Server)(nil)

// Server implements the exchange grpc server.
type Server struct {
	b Backends // TODO(rig): Add reflex streams here.
}

// New returns a new server instance.
func New(b Backends) *Server {
	return &Server{
		b: b,
	}
}

func (srv *Server) Stop() {
}

func (srv *Server) Ping(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	return req, nil
}
