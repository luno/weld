package ops

import "example/sort_with_bind"

type Backends interface {
	GetFoo() sort_with_bind.Foo
	GetBar() sort_with_bind.Bar
	GetBaz() sort_with_bind.Baz
}
