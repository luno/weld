//go:build !dev

package state

// Code generated by weld. DO NOT EDIT.

func MakeBackends() (Backends, error) {
	var (
		b backendsImpl
	)

	return &b, nil
}

type backendsImpl struct {
}
