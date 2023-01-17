package grpc

import (
	"context"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/r4start/goph-keeper/internal/server/app"
	pb "github.com/r4start/goph-keeper/pkg/grpc/proto"
)

const (
	_userAuthKey    = "UserAuth"
	_expectedScheme = "jwt"
)

type AuthService struct {
	pb.UnimplementedAuthorizationServiceServer
	auth             app.Authorizer
	operationTimeout time.Duration
}

func NewAuthService(a app.Authorizer, operationTimeout time.Duration) *AuthService {
	return &AuthService{
		auth:             a,
		operationTimeout: operationTimeout,
	}
}

func (a *AuthService) Register(ctx context.Context, r *pb.AuthorizationRequest) (*pb.AuthorizationResponse, error) {
	if r.Login == nil || r.Password == nil || len(r.Salt) == 0 {
		return nil, status.Error(codes.InvalidArgument, "")
	}

	authCtx, cancel := context.WithTimeout(ctx, a.operationTimeout)
	defer cancel()

	authData, err := a.auth.Register(authCtx, *r.Login, *r.Password, r.Salt)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pb.AuthorizationResponse{
		UserId:       &authData.ID,
		Token:        &authData.Token,
		RefreshToken: &authData.RefreshToken,
	}, nil
}

func (a *AuthService) Authorize(ctx context.Context, r *pb.AuthorizationRequest) (*pb.AuthorizationResponse, error) {
	if r.Login == nil || r.Password == nil {
		return nil, status.Error(codes.InvalidArgument, "")
	}

	authCtx, cancel := context.WithTimeout(ctx, a.operationTimeout)
	defer cancel()

	token, err := a.auth.Authorize(authCtx, *r.Login, *r.Password)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &pb.AuthorizationResponse{
		Token:        &token.Token,
		RefreshToken: &token.RefreshToken,
		UserId:       &token.ID,
	}, nil
}

func (a *AuthService) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	// Since AuthService is responsible for authorization we don't need any middleware to check authorization.
	// Otherwise we won't authorize anybody.
	return ctx, nil
}

func BuildAuthorizationInterceptor(a app.Authorizer) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, _expectedScheme)
		if err != nil {
			return ctx, status.Error(codes.Unauthenticated, "unauthorized")
		}

		auth, err := a.AuthorizeWithToken(ctx, token)
		if err != nil {
			return ctx, status.Error(codes.Unauthenticated, "unauthorized")
		}

		return context.WithValue(ctx, _userAuthKey, auth), nil
	}
}
