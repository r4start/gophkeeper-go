package client

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/r4start/goph-keeper/internal/client/storage"
)

type mockResource struct {
	ID   string
	Salt []byte
	Data []byte
}

type mockClient struct {
	User  storage.UserData
	Files map[string]*mockResource
}

func newMockClient() *mockClient {
	return &mockClient{
		Files: make(map[string]*mockResource),
	}
}

func (m *mockClient) Register(ctx context.Context, login, password string, salt []byte) (*UserAuthorization, error) {
	return nil, nil
}

func (m *mockClient) Authorize(ctx context.Context, login, password string) (*UserAuthorization, error) {
	return nil, nil
}

func (m *mockClient) Store(ctx context.Context, auth *UserAuthorization, salt []byte, fileSize uint64) (ResourceUploader, error) {
	return newMockResourceUploader(m, salt, fileSize), nil
}

func (m *mockClient) List(_ context.Context, _ *UserAuthorization) (RemoteResourcesReader, error) {
	return newMockResourceReader(m), nil
}

func (m *mockClient) Get(ctx context.Context, auth *UserAuthorization, resourceId string) (ResourceDownloader, error) {
	return newMockResourceDownloader(m.Files[resourceId]), nil
}

func (m *mockClient) Delete(ctx context.Context, auth *UserAuthorization, resourceId string) error {
	return nil
}

type mockResourceUploader struct {
	Resource *mockResource
}

func newMockResourceUploader(c *mockClient, salt []byte, fileSize uint64) *mockResourceUploader {
	id, _ := uuid.NewRandom()
	s := make([]byte, len(salt))
	copy(s, salt)
	res := &mockResource{
		ID:   id.String(),
		Salt: s,
		Data: make([]byte, 0, int(fileSize)),
	}
	c.Files[id.String()] = res
	return &mockResourceUploader{
		Resource: res,
	}
}

func (mru *mockResourceUploader) Close() error {
	return nil
}

func (mru *mockResourceUploader) Recv(_ context.Context) (*ResourceInfo, error) {
	return &ResourceInfo{
		ID: mru.Resource.ID,
	}, nil
}

func (mru *mockResourceUploader) SendChunk(_ context.Context, data []byte) error {
	mru.Resource.Data = append(mru.Resource.Data, data...)
	return nil
}

type mockResourceReader struct {
	Resources  []ResourceInfo
	ReadOffset int
}

func newMockResourceReader(client *mockClient) *mockResourceReader {
	res := make([]ResourceInfo, 0, len(client.Files))
	for _, v := range client.Files {
		res = append(res, ResourceInfo{
			ID: v.ID,
		})
	}
	return &mockResourceReader{
		Resources:  res,
		ReadOffset: 0,
	}
}

func (*mockResourceReader) Close() error {
	return nil
}

func (mrr *mockResourceReader) Recv(_ context.Context) (*ResourceInfo, error) {
	if mrr.ReadOffset == len(mrr.Resources) {
		return nil, io.EOF
	}
	pos := mrr.ReadOffset
	mrr.ReadOffset++
	return &mrr.Resources[pos], nil
}

type mockResourceDownloader struct {
	Resource   *mockResource
	ReadOffset int
}

func newMockResourceDownloader(r *mockResource) *mockResourceDownloader {
	return &mockResourceDownloader{
		Resource:   r,
		ReadOffset: -1,
	}
}

func (*mockResourceDownloader) Close() error {
	return nil
}

func (mrd *mockResourceDownloader) Recv(ctx context.Context) (*ResourceChunck, error) {
	if mrd.ReadOffset == -2 {
		return nil, io.EOF
	}
	if mrd.ReadOffset == -1 {
		salt := make([]byte, len(mrd.Resource.Salt))
		copy(salt, mrd.Resource.Salt)
		mrd.ReadOffset = 0
		return &ResourceChunck{
			Salt: salt,
		}, nil
	}
	if mrd.ReadOffset == len(mrd.Resource.Data) {
		size := uint64(mrd.ReadOffset)
		mrd.ReadOffset = -2
		return &ResourceChunck{
			Size: &size,
		}, nil
	}

	data := make([]byte, len(mrd.Resource.Data))
	copy(data, mrd.Resource.Data)
	mrd.ReadOffset = len(mrd.Resource.Data)
	return &ResourceChunck{
		Data: data,
	}, nil
}
