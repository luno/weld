package dev

import (
	"example/exchange"
	"example/exchange/client/grpc"
	"example/exchange/client/logical"
)

func Make(b logical.Backends) (exchange.Client, error) {
	//if !env.IsDev() {
	//	return nil, errors.New("not dev")
	//}
	if grpc.IsEnabled() {
		return grpc.New()
	}

	return logical.New(b), nil
}
