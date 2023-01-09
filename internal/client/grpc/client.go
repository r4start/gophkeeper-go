package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/r4start/goph-keeper/internal/client"
	pb "github.com/r4start/goph-keeper/pkg/grpc/proto"
)

var (
	_ client.Client                = (*grpcClient)(nil)
	_ client.ResourceUploader      = (*grpcResourceUploader)(nil)
	_ client.RemoteResourcesReader = (*grpcRemoteResourceReader)(nil)
	_ client.ResourceDownloader    = (*grpcResourceDownloader)(nil)
)

type grpcClient struct {
	authC    pb.AuthorizationServiceClient
	storageC pb.StorageClient
}

func NewGrpcClient(cfg *client.ServerEndpoint) (*grpcClient, error) {
	var connSecurityOpt grpc.DialOption
	if cfg.UseTLS {
		b, err := os.ReadFile(*cfg.CAPath)
		if err != nil {
			return nil, err
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			return nil, errors.New("credentials: failed to append certificates")
		}
		config := &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            cp,
		}
		connSecurityOpt = grpc.WithTransportCredentials(credentials.NewTLS(config))
	} else {
		connSecurityOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	cc, err := grpc.Dial(cfg.Addr+":"+cfg.Port, connSecurityOpt,
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(16*1024*1024)))
	if err != nil {
		return nil, err
	}
	c := &grpcClient{
		authC:    pb.NewAuthorizationServiceClient(cc),
		storageC: pb.NewStorageClient(cc),
	}
	return c, nil
}

func (g *grpcClient) Register(ctx context.Context, login, password string, salt []byte) (*client.UserAuthorization, error) {
	auth, err := g.authC.Register(ctx, &pb.AuthorizationRequest{
		Login:    &login,
		Password: &password,
		Salt:     salt,
	})
	if err != nil {
		return nil, err
	}

	result := &client.UserAuthorization{}
	result.Token = *auth.Token
	result.RefreshToken = *auth.RefreshToken
	result.UserID = *auth.UserId

	return result, nil
}

func (g *grpcClient) Authorize(ctx context.Context, login, password string) (*client.UserAuthorization, error) {
	auth, err := g.authC.Authorize(ctx, &pb.AuthorizationRequest{
		Login:    &login,
		Password: &password,
	})
	if err != nil {
		return nil, err
	}

	result := &client.UserAuthorization{}
	result.Token = *auth.Token
	result.RefreshToken = *auth.RefreshToken
	result.UserID = *auth.UserId
	result.Salt = auth.Salt

	return result, nil
}

func (g *grpcClient) Store(ctx context.Context, auth *client.UserAuthorization, salt []byte, fileSize uint64) (client.ResourceUploader, error) {
	rctx := addAuth(ctx, auth)
	streamingC, err := g.storageC.Add(rctx)
	if err != nil {
		return nil, err
	}

	if err := streamingC.Send(&pb.ResourceOperationData{
		Data: &pb.ResourceOperationData_Meta{Meta: &pb.ResourceOperationData_ResourceMeta{
			Salt:             salt,
			ResourceByteSize: &fileSize,
		}},
	}); err != nil {
		return nil, err
	}

	return &grpcResourceUploader{sC: streamingC}, nil
}

func (g *grpcClient) List(ctx context.Context, auth *client.UserAuthorization) (client.RemoteResourcesReader, error) {
	rctx := addAuth(ctx, auth)
	req := &pb.ListRequest{}
	listC, err := g.storageC.List(rctx, req)
	if err != nil {
		return nil, err
	}
	return &grpcRemoteResourceReader{lC: listC}, nil
}

func (g *grpcClient) Get(ctx context.Context, auth *client.UserAuthorization, resourceId string) (client.ResourceDownloader, error) {
	rctx := addAuth(ctx, auth)
	c, err := g.storageC.Get(rctx, &pb.Resource{
		Id: &resourceId,
	})
	if err != nil {
		return nil, err
	}
	return &grpcResourceDownloader{
		C: c,
	}, nil
}

func (g *grpcClient) Delete(ctx context.Context, auth *client.UserAuthorization, resourceId string) error {
	rctx := addAuth(ctx, auth)
	_, err := g.storageC.Delete(rctx, &pb.Resource{
		Id: &resourceId,
	})
	return err
}

func addAuth(c context.Context, auth *client.UserAuthorization) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("jwt %v", auth.Token))
	return metautils.NiceMD(md).ToOutgoing(c)
}

type grpcResourceUploader struct {
	sC pb.Storage_AddClient
}

func (g *grpcResourceUploader) Close() error {
	return g.sC.CloseSend()
}

func (g *grpcResourceUploader) SendChunk(_ context.Context, data []byte) error {
	m := &pb.ResourceOperationData{
		Data: &pb.ResourceOperationData_Chunk{
			Chunk: &pb.ResourceOperationData_DataChunk{
				Data: data,
			},
		},
	}
	if err := g.sC.Send(m); err != nil {
		return err
	}
	return nil
}

func (g *grpcResourceUploader) Recv(_ context.Context) (*client.ResourceInfo, error) {
	m, err := g.sC.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	errCode := pb.ErrorCode(m.GetErrorCode())
	if errCode != pb.ErrorCode_ERROR_CODE_OK {
		return nil, fmt.Errorf("got an error from the server: %s", errCode.String())
	}
	info := &client.ResourceInfo{
		ID: *m.GetResource().Id,
	}
	return info, err
}

type grpcRemoteResourceReader struct {
	lC pb.Storage_ListClient
}

func (r *grpcRemoteResourceReader) Close() error {
	return r.lC.CloseSend()
}

func (r *grpcRemoteResourceReader) Recv(ctx context.Context) (*client.ResourceInfo, error) {
	resource, err := r.lC.Recv()
	if err != nil {
		return nil, err
	}
	return &client.ResourceInfo{
		ID: *resource.Id,
	}, nil
}

type grpcResourceDownloader struct {
	C pb.Storage_GetClient
}

// Close implements client.ResourceDownloader
func (g *grpcResourceDownloader) Close() error {
	return g.C.CloseSend()
}

// Recv implements client.ResourceDownloader
func (g *grpcResourceDownloader) Recv(context.Context) (*client.ResourceChunck, error) {
	m, err := g.C.Recv()
	if err != nil {
		return nil, err
	}

	result := &client.ResourceChunck{}
	if meta := m.GetMeta(); meta != nil {
		result.Salt = meta.Salt
		result.Size = meta.ResourceByteSize
	}

	if chunk := m.GetChunk(); chunk != nil {
		result.Data = chunk.GetData()
	}

	return result, nil
}
