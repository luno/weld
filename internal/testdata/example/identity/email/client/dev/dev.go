package dev

import (
	"example/identity/email"
	"example/identity/email/client/grpc"
	"example/identity/email/client/logical"
)

func Make(b logical.Backends) (email.Client, error) {
	//if !env.IsDev() {
	//	return nil, errors.New("not dev")
	//}
	if grpc.IsEnabled() {
		return grpc.New()
	}

	return logical.New(b), nil
}
