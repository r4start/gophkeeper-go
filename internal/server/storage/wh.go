package storage

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type ResourceID uuid.UUID

func (r ResourceID) String() string {
	return uuid.UUID(r).String()
}

type Resource interface {
	io.Closer
	io.Writer
	io.Reader

	GetId() *ResourceID
	IsDeleted() bool
	Salt() ([]byte, error)
}

type Storage interface {
	Create(ctx context.Context, user *UserID, salt []byte) (Resource, error)
	Open(ctx context.Context, user *UserID, id *ResourceID) (Resource, error)
	Delete(ctx context.Context, user *UserID, id *ResourceID) error
	List(ctx context.Context, user *UserID) ([]Resource, error)
}
