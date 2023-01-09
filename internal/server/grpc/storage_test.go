package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"sort"
	"testing"

	"github.com/google/uuid"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	pb "github.com/r4start/goph-keeper/pkg/grpc/proto"

	"github.com/r4start/goph-keeper/internal/server/app"
	"github.com/r4start/goph-keeper/internal/server/storage"
)

func generateRandom(size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func authGenerator(id string) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		auth := &app.AuthData{
			ID: id,
		}
		return context.WithValue(ctx, userAuthKey, auth), nil
	}
}

type mockWhStorage struct {
	Resources map[uuid.UUID]*mockResource
}

func newMockWhStorage() *mockWhStorage {
	return &mockWhStorage{
		Resources: make(map[uuid.UUID]*mockResource),
	}
}

func (m *mockWhStorage) Create(ctx context.Context, _ *storage.UserID, salt []byte) (storage.Resource, error) {
	resID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	res := &mockResource{
		ID:         resID,
		Buffer:     make([]byte, 0),
		ReadOffset: 0,
		SaltData:   salt,
	}

	m.Resources[resID] = res

	return res, nil
}

func (m *mockWhStorage) Open(ctx context.Context, user *storage.UserID, id *storage.ResourceID) (storage.Resource, error) {
	res, ok := m.Resources[uuid.UUID(*id)]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return res, nil
}

func (m *mockWhStorage) Delete(ctx context.Context, user *storage.UserID, id *storage.ResourceID) error {
	if _, ok := m.Resources[uuid.UUID(*id)]; !ok {
		return fmt.Errorf("not found")
	}
	delete(m.Resources, uuid.UUID(*id))
	return nil
}

func (m *mockWhStorage) List(ctx context.Context, user *storage.UserID) ([]storage.Resource, error) {
	result := make([]storage.Resource, 0, len(m.Resources))
	for _, v := range m.Resources {
		result = append(result, v)
	}
	return result, nil
}

type mockResource struct {
	ID         uuid.UUID
	Buffer     []byte
	ReadOffset int
	SaltData   []byte
}

func (mr *mockResource) Close() error {
	mr.ReadOffset = 0
	return nil
}

func (mr *mockResource) Write(p []byte) (n int, err error) {
	mr.Buffer = append(mr.Buffer, p...)
	return len(p), nil
}

func (mr *mockResource) Read(p []byte) (n int, err error) {
	bufSize := len(mr.Buffer)
	if mr.ReadOffset == bufSize {
		return 0, io.EOF
	}

	bufLen := len(p)
	if mr.ReadOffset+bufLen > bufSize {
		bufLen = len(mr.Buffer) - mr.ReadOffset
	}
	copy(p, mr.Buffer[mr.ReadOffset:mr.ReadOffset+bufLen])
	mr.ReadOffset += bufLen
	return bufLen, nil
}

func (mr *mockResource) GetId() *storage.ResourceID {
	uid := storage.ResourceID(mr.ID)
	return &uid
}

func (mr *mockResource) IsDeleted() bool {
	return false
}

func (mr *mockResource) Salt() ([]byte, error) {
	return mr.SaltData, nil
}

func TestStorageService_Add(t *testing.T) {
	userID, err := uuid.NewRandom()
	assert.NoError(t, err)
	s, err := NewStorageService(newMockWhStorage())
	assert.NoError(t, err)

	reg := func(srv *grpc.Server) {
		pb.RegisterStorageServer(srv, s)
	}

	authFunc := authGenerator(userID.String())

	ctx := context.Background()
	srv, conn := prepareTestEnv(t, reg,
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)))
	defer srv.Stop()
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewStorageClient(conn)

	streamC, err := client.Add(ctx)
	assert.NoError(t, err)

	salt, err := generateRandom(64)
	assert.NoError(t, err)

	data, err := generateRandom(50 * 1024 * 1024)
	assert.NoError(t, err)

	size := uint64(len(data))
	err = streamC.Send(&pb.ResourceOperationData{
		Data: &pb.ResourceOperationData_Meta{Meta: &pb.ResourceOperationData_ResourceMeta{
			Salt:             salt,
			ResourceByteSize: &size,
		}},
	})
	assert.NoError(t, err)

	bufferSize := uint64(1024)
	for offset := uint64(0); offset < size; offset += bufferSize {
		m := &pb.ResourceOperationData{
			Data: &pb.ResourceOperationData_Chunk{
				Chunk: &pb.ResourceOperationData_DataChunk{
					Data: data[offset : offset+bufferSize],
				},
			},
		}
		assert.NoError(t, streamC.Send(m))
	}

	m, err := streamC.CloseAndRecv()
	assert.NoError(t, err)

	errCode := pb.ErrorCode(m.GetErrorCode())
	assert.Equal(t, pb.ErrorCode_ERROR_CODE_OK, errCode)
	assert.NotNil(t, m.GetResource().Id)
	assert.NotZero(t, len(*m.GetResource().Id))
}

func TestStorageService_Add_WrongStart(t *testing.T) {
	userID, err := uuid.NewRandom()
	assert.NoError(t, err)
	s, err := NewStorageService(newMockWhStorage())
	assert.NoError(t, err)

	reg := func(srv *grpc.Server) {
		pb.RegisterStorageServer(srv, s)
	}

	authFunc := authGenerator(userID.String())

	ctx := context.Background()
	srv, conn := prepareTestEnv(t, reg,
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)))
	defer srv.Stop()
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewStorageClient(conn)

	streamC, err := client.Add(ctx)
	assert.NoError(t, err)

	data, err := generateRandom(1024)
	assert.NoError(t, err)

	m := &pb.ResourceOperationData{
		Data: &pb.ResourceOperationData_Chunk{
			Chunk: &pb.ResourceOperationData_DataChunk{
				Data: data,
			},
		},
	}
	assert.NoError(t, streamC.Send(m))

	_, err = streamC.CloseAndRecv()
	assert.Error(t, err)
}

func TestStorageService_List(t *testing.T) {
	userID, err := uuid.NewRandom()
	assert.NoError(t, err)
	s, err := NewStorageService(newMockWhStorage())
	assert.NoError(t, err)

	reg := func(srv *grpc.Server) {
		pb.RegisterStorageServer(srv, s)
	}

	authFunc := authGenerator(userID.String())

	ctx := context.Background()
	srv, conn := prepareTestEnv(t, reg,
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)))
	defer srv.Stop()
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewStorageClient(conn)

	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		streamC, err := client.Add(ctx)
		assert.NoError(t, err)

		salt, err := generateRandom(64)
		assert.NoError(t, err)

		data, err := generateRandom(5 * 1024 * 1024)
		assert.NoError(t, err)

		size := uint64(len(data))
		err = streamC.Send(&pb.ResourceOperationData{
			Data: &pb.ResourceOperationData_Meta{Meta: &pb.ResourceOperationData_ResourceMeta{
				Salt:             salt,
				ResourceByteSize: &size,
			}},
		})
		assert.NoError(t, err)

		bufferSize := uint64(1024)
		for offset := uint64(0); offset < size; offset += bufferSize {
			m := &pb.ResourceOperationData{
				Data: &pb.ResourceOperationData_Chunk{
					Chunk: &pb.ResourceOperationData_DataChunk{
						Data: data[offset : offset+bufferSize],
					},
				},
			}
			assert.NoError(t, streamC.Send(m))
		}

		m, err := streamC.CloseAndRecv()
		assert.NoError(t, err)

		errCode := pb.ErrorCode(m.GetErrorCode())
		assert.Equal(t, pb.ErrorCode_ERROR_CODE_OK, errCode)
		assert.NotNil(t, m.GetResource().Id)
		assert.NotZero(t, len(*m.GetResource().Id))
		ids[i] = *m.GetResource().Id
	}

	listC, err := client.List(ctx, &pb.ListRequest{})
	assert.NoError(t, err)
	remoteResources := make([]string, 0, len(ids))
	for {
		rr, err := listC.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		remoteResources = append(remoteResources, *rr.Id)
	}

	sort.Strings(ids)
	sort.Strings(remoteResources)
	assert.Equal(t, ids, remoteResources)
}

func TestStorageService_Get(t *testing.T) {
	storage := newMockWhStorage()

	userID, err := uuid.NewRandom()
	assert.NoError(t, err)
	s, err := NewStorageService(storage)
	assert.NoError(t, err)

	reg := func(srv *grpc.Server) {
		pb.RegisterStorageServer(srv, s)
	}

	authFunc := authGenerator(userID.String())

	ctx := context.Background()
	srv, conn := prepareTestEnv(t, reg,
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)))
	defer srv.Stop()
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewStorageClient(conn)

	streamC, err := client.Add(ctx)
	assert.NoError(t, err)

	salt, err := generateRandom(64)
	assert.NoError(t, err)

	data, err := generateRandom(50 * 1024 * 1024)
	assert.NoError(t, err)

	size := uint64(len(data))
	err = streamC.Send(&pb.ResourceOperationData{
		Data: &pb.ResourceOperationData_Meta{Meta: &pb.ResourceOperationData_ResourceMeta{
			Salt:             salt,
			ResourceByteSize: &size,
		}},
	})
	assert.NoError(t, err)

	bufferSize := uint64(1024)
	for offset := uint64(0); offset < size; offset += bufferSize {
		m := &pb.ResourceOperationData{
			Data: &pb.ResourceOperationData_Chunk{
				Chunk: &pb.ResourceOperationData_DataChunk{
					Data: data[offset : offset+bufferSize],
				},
			},
		}
		assert.NoError(t, streamC.Send(m))
	}

	m, err := streamC.CloseAndRecv()
	assert.NoError(t, err)

	errCode := pb.ErrorCode(m.GetErrorCode())
	assert.Equal(t, pb.ErrorCode_ERROR_CODE_OK, errCode)
	assert.NotNil(t, m.GetResource().Id)
	assert.NotZero(t, len(*m.GetResource().Id))

	getC, err := client.Get(ctx, &pb.Resource{
		Id: m.GetResource().Id,
	})
	assert.NoError(t, err)

	var (
		remoteSize uint64
		remoteSalt []byte
		remoteData = make([]byte, 0, len(data))
	)
	for {
		chunk, err := getC.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)

		if meta := chunk.GetMeta(); meta != nil {
			if meta.Salt != nil {
				assert.Nil(t, remoteSalt)
				assert.Nil(t, chunk.GetChunk())

				remoteSalt = meta.Salt
			}
			if meta.ResourceByteSize != nil {
				assert.NotNil(t, remoteSalt)
				assert.Zero(t, remoteSize)
				assert.Nil(t, chunk.GetChunk())
				remoteSize = *meta.ResourceByteSize
			}

			continue
		}

		assert.NotNil(t, chunk.GetChunk())
		assert.NotNil(t, chunk.GetChunk().GetData())

		remoteData = append(remoteData, chunk.GetChunk().GetData()...)
	}

	assert.Equal(t, salt, remoteSalt)
	assert.Equal(t, size, uint64(len(remoteData)))
	assert.Equal(t, size, remoteSize)
	assert.Equal(t, data, remoteData)

	rndRes, err := uuid.NewRandom()
	assert.NoError(t, err)

	_, ok := storage.Resources[rndRes]
	assert.False(t, ok)

	testRes := rndRes.String()
	getC, err = client.Get(ctx, &pb.Resource{
		Id: &testRes,
	})
	assert.NoError(t, err)
	_, err = getC.Recv()
	assert.Error(t, err)
}

func TestStorageService_Delete(t *testing.T) {
	storage := newMockWhStorage()
	userID, err := uuid.NewRandom()
	assert.NoError(t, err)
	s, err := NewStorageService(storage)
	assert.NoError(t, err)

	reg := func(srv *grpc.Server) {
		pb.RegisterStorageServer(srv, s)
	}

	authFunc := authGenerator(userID.String())

	ctx := context.Background()
	srv, conn := prepareTestEnv(t, reg,
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)))
	defer srv.Stop()
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewStorageClient(conn)

	ids := make([]string, 10)
	for i := 0; i < len(ids); i++ {
		streamC, err := client.Add(ctx)
		assert.NoError(t, err)

		salt, err := generateRandom(64)
		assert.NoError(t, err)

		data, err := generateRandom(5 * 1024 * 1024)
		assert.NoError(t, err)

		size := uint64(len(data))
		err = streamC.Send(&pb.ResourceOperationData{
			Data: &pb.ResourceOperationData_Meta{Meta: &pb.ResourceOperationData_ResourceMeta{
				Salt:             salt,
				ResourceByteSize: &size,
			}},
		})
		assert.NoError(t, err)

		bufferSize := uint64(1024)
		for offset := uint64(0); offset < size; offset += bufferSize {
			m := &pb.ResourceOperationData{
				Data: &pb.ResourceOperationData_Chunk{
					Chunk: &pb.ResourceOperationData_DataChunk{
						Data: data[offset : offset+bufferSize],
					},
				},
			}
			assert.NoError(t, streamC.Send(m))
		}

		m, err := streamC.CloseAndRecv()
		assert.NoError(t, err)

		errCode := pb.ErrorCode(m.GetErrorCode())
		assert.Equal(t, pb.ErrorCode_ERROR_CODE_OK, errCode)
		assert.NotNil(t, m.GetResource().Id)
		assert.NotZero(t, len(*m.GetResource().Id))
		ids[i] = *m.GetResource().Id
	}

	assert.Equal(t, len(ids), len(storage.Resources))

	for i := 0; i < len(ids); i++ {
		_, err := client.Delete(ctx, &pb.Resource{
			Id: &ids[i],
		})
		assert.NoError(t, err)
		id, err := uuid.Parse(ids[i])
		assert.NoError(t, err)
		_, ok := storage.Resources[id]
		assert.False(t, ok)
	}
	assert.Zero(t, len(storage.Resources))

	rndRes, err := uuid.NewRandom()
	assert.NoError(t, err)
	rndStr := rndRes.String()
	_, err = client.Delete(ctx, &pb.Resource{
		Id: &rndStr,
	})
	assert.Error(t, err)
}
