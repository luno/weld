package versioned

import (
	versioned "example/external/versioned/v1"
)

func New() *versioned.Service {
	return new(versioned.Service)
}
