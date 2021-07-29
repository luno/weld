package dev

import (
	"example/identity/users"
	"example/identity/users/client/grpc"
	"example/identity/users/client/logical"
)

func Make(b logical.Backends) (users.Client, error) {
	//if !env.IsDev() {
	//	return nil, errors.New("not dev")
	//}
	if grpc.IsEnabled() {
		return grpc.New()
	}

	return logical.New(b), nil
}
