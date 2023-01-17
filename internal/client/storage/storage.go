package storage

import (
	"context"

	"github.com/r4start/goph-keeper/internal/crypto"
)

type Storage interface {
	UserData(ctx context.Context) (*UserData, error)
	SetUserData(ctx context.Context, ud *UserData) error

	AddFile(ctx context.Context, fd *FileData) error
	ListFiles(ctx context.Context) ([]FileData, error)
	DeleteFile(ctx context.Context, fd *FileData) error
	FileData(ctx context.Context, id string) (*FileData, error)

	AddCard(ctx context.Context, data *CardData) error
	ListCards(ctx context.Context) ([]CardData, error)
	DeleteCard(ctx context.Context, id string) error
	CardData(ctx context.Context, id string) (*CardData, error)

	AddCredentials(ctx context.Context, data *CredentialData) error
	ListCredentials(ctx context.Context) ([]CredentialData, error)
	DeleteCredential(ctx context.Context, id string) error
	CredentialData(ctx context.Context, id string) (*CredentialData, error)
}

type UserData struct {
	UserID       string
	Token        string
	RefreshToken string
	MasterKey    []byte
	Salt         []byte
}

type FileData struct {
	ID     string
	UserID string
	Name   string
	Path   string
	Key    crypto.Secret
}

type CardData struct {
	ID           string
	UserID       string
	Name         string
	Number       string
	Holder       string
	ExpiryDate   string
	SecurityCode string
}

type CredentialData struct {
	ID          string
	UserID      string
	Username    string
	Password    string
	Uri         string
	Description string
}
