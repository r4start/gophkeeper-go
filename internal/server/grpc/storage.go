package grpc

import (
	"context"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/r4start/goph-keeper/internal/server/app"
	"github.com/r4start/goph-keeper/internal/server/storage"
	pb "github.com/r4start/goph-keeper/pkg/grpc/proto"
)

type StorageService struct {
	pb.UnimplementedStorageServer

	wh             storage.Storage
	sendBufferSize int
}

func NewStorageService(wh storage.Storage, sendBufferSize int) (*StorageService, error) {
	return &StorageService{wh: wh, sendBufferSize: sendBufferSize}, nil
}

func (s *StorageService) Add(stream pb.Storage_AddServer) error {
	var (
		res          storage.Resource
		expectedSize = uint64(0)
		readBytes    = uint64(0)
		ctx          = stream.Context()
		userAuth, ok = ctx.Value(_userAuthKey).(*app.AuthData)
	)

	if !ok || userAuth == nil {
		return status.Error(codes.Unauthenticated, "auth token missed")
	}

	userID, err := storage.NewUserIDFromString(userAuth.ID)
	if err != nil {
		return status.Error(codes.Unauthenticated, "bad user id")
	}

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			if res == nil {
				return status.Error(codes.FailedPrecondition, "must start with meta information")
			}
			id := res.GetId().String()
			if err := res.Close(); err != nil {
				return err
			}

			return stream.SendAndClose(&pb.ResourceOperationResponse{
				Result: &pb.ResourceOperationResponse_Resource{
					Resource: &pb.Resource{
						Id: &id,
					},
				},
			})
		}
		if err != nil {
			return err
		}
		switch v := data.GetData().(type) {
		case *pb.ResourceOperationData_Meta:
			if res, err = s.wh.Create(ctx, userID, v.Meta.Salt); err != nil {
				return err
			}
			if v.Meta.ResourceByteSize == nil {
				return status.Errorf(codes.InvalidArgument, "resource byte size must be specified")
			}
			expectedSize = *v.Meta.ResourceByteSize

		case *pb.ResourceOperationData_Chunk:
			if res == nil {
				return status.Error(codes.FailedPrecondition, "must start with meta information")
			}

			n, err := res.Write(v.Chunk.Data)
			if err != nil {
				return err
			}

			if n != len(v.Chunk.Data) {
				return status.Error(codes.ResourceExhausted, "failed to save data")
			}

			readBytes += uint64(len(v.Chunk.Data))
			if readBytes > expectedSize {
				return status.Error(codes.OutOfRange, "data is larger than expected")
			}
		}
	}
}

func (s *StorageService) List(_ *pb.ListRequest, stream pb.Storage_ListServer) error {
	var (
		ctx          = stream.Context()
		userAuth, ok = ctx.Value(_userAuthKey).(*app.AuthData)
	)

	if !ok || userAuth == nil {
		return status.Error(codes.Unauthenticated, "auth token missed")
	}

	userID, err := storage.NewUserIDFromString(userAuth.ID)
	if err != nil {
		return status.Error(codes.Unauthenticated, "bad user id")
	}

	resources, err := s.wh.List(ctx, userID)
	if err != nil {
		return err
	}

	for _, r := range resources {
		id := r.GetId().String()
		err := stream.Send(&pb.Resource{
			Id: &id,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StorageService) Get(res *pb.Resource, stream pb.Storage_GetServer) error {
	if res.Id == nil {
		return status.Errorf(codes.InvalidArgument, "resource id is empty")
	}

	id, err := uuid.Parse(*res.Id)
	if err != nil {
		return err
	}

	ctx := stream.Context()
	userAuth, ok := ctx.Value(_userAuthKey).(*app.AuthData)
	if !ok || userAuth == nil {
		return status.Error(codes.Unauthenticated, "auth token missed")
	}

	userID, err := storage.NewUserIDFromString(userAuth.ID)
	if err != nil {
		return status.Error(codes.Unauthenticated, "bad user id")
	}

	resId := storage.ResourceID(id)
	resource, err := s.wh.Open(ctx, userID, &resId)
	if err != nil {
		return err
	}
	defer func() {
		_ = resource.Close()
	}()

	salt, err := resource.Salt()
	if err != nil {
		return err
	}

	err = stream.Send(&pb.ResourceOperationData{
		Data: &pb.ResourceOperationData_Meta{
			Meta: &pb.ResourceOperationData_ResourceMeta{
				Salt: salt,
			},
		},
	})
	if err != nil {
		return err
	}

	readSize := uint64(0)
	buffer := make([]byte, s.sendBufferSize)
	for {
		readBytes, err := resource.Read(buffer)
		if err == io.EOF {
			if readBytes != 0 {
				readSize += uint64(readBytes)
				err = stream.Send(&pb.ResourceOperationData{
					Data: &pb.ResourceOperationData_Chunk{
						Chunk: &pb.ResourceOperationData_DataChunk{
							Data: buffer[:readBytes],
						},
					},
				})
				if err != nil {
					return err
				}
			}
			break
		}
		if err != nil {
			return err
		}

		readSize += uint64(readBytes)
		err = stream.Send(&pb.ResourceOperationData{
			Data: &pb.ResourceOperationData_Chunk{
				Chunk: &pb.ResourceOperationData_DataChunk{
					Data: buffer[:readBytes],
				},
			},
		})
		if err != nil {
			return err
		}
	}

	err = stream.Send(&pb.ResourceOperationData{
		Data: &pb.ResourceOperationData_Meta{
			Meta: &pb.ResourceOperationData_ResourceMeta{
				ResourceByteSize: &readSize,
			},
		},
	})
	return err
}

func (s *StorageService) Delete(ctx context.Context, res *pb.Resource) (*pb.ResourceOperationResponse, error) {
	if res.Id == nil {
		return nil, status.Errorf(codes.InvalidArgument, "resource id is empty")
	}

	id, err := uuid.Parse(*res.Id)
	if err != nil {
		return nil, err
	}

	userAuth, ok := ctx.Value(_userAuthKey).(*app.AuthData)
	if !ok || userAuth == nil {
		return nil, status.Error(codes.Unauthenticated, "auth token missed")
	}

	userId, err := storage.NewUserIDFromString(userAuth.ID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "bad user id")
	}

	resId := storage.ResourceID(id)
	if err := s.wh.Delete(ctx, userId, &resId); err != nil {
		return nil, err
	}

	return &pb.ResourceOperationResponse{
		Result: &pb.ResourceOperationResponse_ErrorCode{
			ErrorCode: 0,
		},
	}, nil
}
