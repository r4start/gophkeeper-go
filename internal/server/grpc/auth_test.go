package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/stretchr/testify/assert"

	"github.com/r4start/goph-keeper/internal/server/app"
	pb "github.com/r4start/goph-keeper/pkg/grpc/proto"
)

type mockAuth struct {
	users *sync.Map
}

func (s *mockAuth) Register(_ context.Context, login, password string, keySalt []byte) (*app.AuthData, error) {
	defaultToken := "AAAAAAAAA"

	if len(login) == 0 || len(password) == 0 || len(keySalt) == 0 {
		return nil, fmt.Errorf("constraints check failed")
	}

	_, loaded := s.users.LoadOrStore(login, password)
	if loaded {
		return nil, fmt.Errorf("user exists")
	}

	return &app.AuthData{
		Token:        defaultToken,
		RefreshToken: defaultToken,
		ExpiresAt:    0,
	}, nil
}

func (s *mockAuth) Authorize(_ context.Context, login, password string) (*app.AuthData, error) {
	defaultToken := "AAAAAAAAA"

	if len(login) == 0 || len(password) == 0 {
		return nil, fmt.Errorf("constraints check failed")
	}

	pwd, loaded := s.users.Load(login)
	if !loaded {
		return nil, fmt.Errorf("user doesn't exist")
	}

	if pwd != password {
		return nil, fmt.Errorf("authorization failed")
	}

	return &app.AuthData{
		Token:        defaultToken,
		RefreshToken: defaultToken,
		ExpiresAt:    0,
	}, nil
}

func (s *mockAuth) RefreshToken(ctx context.Context, refreshToken string) (*app.AuthData, error) {
	panic("unimplemented")
}

func (s *mockAuth) IsValidToken(ctx context.Context, token string) error {
	panic("unimplemented")
}

func (s *mockAuth) AuthorizeWithToken(ctx context.Context, token string) (*app.AuthData, error) {
	panic("unimplemented")
}

type serviceRegisterer func(srv *grpc.Server)

func prepareServer(t *testing.T, register serviceRegisterer, opts ...grpc.ServerOption) (*grpc.Server, *bufconn.Listener) {
	const bufSize = 1024 * 1024

	lis := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer(opts...)
	register(grpcServer)
	go func(t *testing.T) {
		err := grpcServer.Serve(lis)
		assert.NoError(t, err)
	}(t)

	return grpcServer, lis
}

func makeDialer(l *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	}
}

func prepareTestEnv(t *testing.T, reg serviceRegisterer, opts ...grpc.ServerOption) (*grpc.Server, *grpc.ClientConn) {
	grpcServer, listener := prepareServer(t, reg, opts...)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(makeDialer(listener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return grpcServer, conn
}

func strPtr(s string) *string {
	r := s
	return &r
}

func TestAuthService_Basic(t *testing.T) {
	keySalt := make([]byte, 64)

	s := NewAuthService(&mockAuth{users: new(sync.Map)}, time.Second)
	reg := func(srv *grpc.Server) {
		pb.RegisterAuthorizationServiceServer(srv, s)
	}

	srv, conn := prepareTestEnv(t, reg)
	defer srv.Stop()
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewAuthorizationServiceClient(conn)
	ctx := context.Background()

	test1 := &pb.AuthorizationRequest{
		Login:    strPtr("test1"),
		Password: strPtr("test1"),
		Salt:     keySalt,
	}

	resp, err := client.Register(ctx, test1)
	assert.NoError(t, err)
	assert.NotZero(t, len(*resp.Token))

	resp, err = client.Authorize(ctx, test1)
	assert.NoError(t, err)
	assert.NotZero(t, len(*resp.Token))

	_, err = client.Register(ctx, test1)
	assert.Error(t, err)

	test1.Password = strPtr("1test")
	_, err = client.Authorize(ctx, test1)
	assert.Error(t, err)

	test2 := &pb.AuthorizationRequest{
		Login:    strPtr("test2"),
		Password: strPtr("test1"),
		Salt:     keySalt,
	}
	_, err = client.Authorize(ctx, test2)
	assert.Error(t, err)

	test3 := &pb.AuthorizationRequest{
		Login:    strPtr(""),
		Password: strPtr("test1"),
		Salt:     keySalt,
	}
	_, err = client.Register(ctx, test3)
	assert.Error(t, err)

	test4 := &pb.AuthorizationRequest{
		Login:    strPtr("test4"),
		Password: strPtr(""),
		Salt:     keySalt,
	}
	_, err = client.Register(ctx, test4)
	assert.Error(t, err)
}
