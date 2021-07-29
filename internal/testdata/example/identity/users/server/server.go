package server

import (
	"context"
	"example/identity/users/ops"
	pb "example/identity/users/userspb"
)

var _ pb.UsersServer = (*Server)(nil)

// Server implements the users grpc server.
type Server struct {
	b adapter
}

// New returns a new server instance.
func New(b Backends) *Server {
	return &Server{
		b: adapter{Backends: b},
	}
}

func (srv *Server) Stop() {
}

func (srv *Server) Ping(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	if err := ops.Logic(srv.b); err != nil {
		return nil, err
	}
	return req, nil
}
