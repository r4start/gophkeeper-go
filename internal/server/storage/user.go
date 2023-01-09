package storage

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type UserID uuid.UUID

func (u UserID) String() string {
	return uuid.UUID(u).String()
}

func NewUserIDFromString(u string) (*UserID, error) {
	id, err := uuid.Parse(u)
	if err != nil {
		return nil, err
	}
	res := UserID(id)
	return &res, nil
}

type User struct {
	ID        UserID
	Login     string
	KeySalt   []byte
	Salt      []byte
	Secret    []byte
	IsDeleted bool
}

type UserService interface {
	Add(ctx context.Context, login string, keySalt, salt, secret []byte) (*UserID, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)

	io.Closer
}
