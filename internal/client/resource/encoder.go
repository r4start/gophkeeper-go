package resource

import (
	"context"
)

type CodingCapabilitiesObserver interface {
	Key() []byte
	Salt() []byte
}

type Encoder interface {
	CodingCapabilitiesObserver
	Encode(ctx context.Context, data []byte) ([]byte, error)
}

type Decoder interface {
	CodingCapabilitiesObserver
	Decode(ctx context.Context, data []byte) ([]byte, error)
}
