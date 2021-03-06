package state

// Code generated by weld. DO NOT EDIT.

import (
	"example/samevar/pool"
)

func MakeBackends() (Backends, error) {
	var (
		b backendsImpl
	)

	b.pool = pool.NewBarPool()

	b.pool1 = pool.NewFooPool()

	return &b, nil
}

type backendsImpl struct {
	pool  *pool.BarPool
	pool1 *pool.FooPool
}

func (b *backendsImpl) GetPool() *pool.BarPool {
	return b.pool
}

func (b *backendsImpl) Pool() *pool.FooPool {
	return b.pool1
}
