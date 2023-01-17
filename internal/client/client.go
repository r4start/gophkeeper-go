package client

import (
	"context"
	"io"
)

type Client interface {
	Register(ctx context.Context, login, password string, salt []byte) (*UserAuthorization, error)
	Authorize(ctx context.Context, login, password string) (*UserAuthorization, error)

	Store(ctx context.Context, auth *UserAuthorization, salt []byte, fileSize uint64) (ResourceUploader, error)
	List(ctx context.Context, auth *UserAuthorization) (RemoteResourcesReader, error)
	Get(ctx context.Context, auth *UserAuthorization, resourceId string) (ResourceDownloader, error)
	Delete(ctx context.Context, auth *UserAuthorization, resourceId string) error
}

type ResourceDownloader interface {
	io.Closer

	Recv(ctx context.Context) (*ResourceChunck, error)
}

type RemoteResourcesReader interface {
	io.Closer

	Recv(ctx context.Context) (*ResourceInfo, error)
}

type ServerEndpoint struct {
	Addr   string  `json:"address"`
	Port   string  `json:"port"`
	UseTLS bool    `json:"use_tls"`
	CAPath *string `json:"ca_path,omitempty"`
}

type UserAuthorization struct {
	Token        string
	RefreshToken string
	UserID       string
	Salt         []byte
}

type ResourceInfo struct {
	ErrorCode int32
	ID        string
	IsDeleted bool
}

type ResourceUploader interface {
	io.Closer

	SendChunk(ctx context.Context, data []byte) error
	Recv(ctx context.Context) (*ResourceInfo, error)
}

type ResourceChunck struct {
	Salt []byte
	Data []byte
	Size *uint64
}
