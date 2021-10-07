package ops1

import "example/samevar/pool"

type Backends interface {
	Pool() *pool.FooPool
}
