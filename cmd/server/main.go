package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/flags"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"go.uber.org/zap"

	"github.com/r4start/goph-keeper/internal/server/app"
	gsrv "github.com/r4start/goph-keeper/internal/server/grpc"
	"github.com/r4start/goph-keeper/internal/server/storage"
	pb "github.com/r4start/goph-keeper/pkg/grpc/proto"
)

type config struct {
	DatabaseConnectionString string `config:"db_dsn,required"`
	DatabaseOperationTimeout uint32 `config:"db_timeout"`
	TokenSignKeyFilePath     string `config:"token_key,required"`
	GrpcServerAddress        string `config:"grpc_server_address"`
	GrpcServerBasePort       uint16 `config:"grpc_server_base_port"`
	GrpcServerRecvSize       int    `config:"grpc_server_recv_size"`
	GrpcServerSendSize       int    `config:"grpc_server_send_size"`
	ServeTLS                 bool   `config:"use_tls"`
	TLSKeyFilePath           string `config:"key_file"`
	TLSCrtFilePath           string `config:"crt_file"`
	RPSLimit                 uint32 `config:"rps_limit"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("failed to initialize logger: %+v", err)
		return
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Println(err)
		}
	}()

	cfg := &config{
		DatabaseOperationTimeout: 250,
		GrpcServerBasePort:       8090,
		GrpcServerRecvSize:       16 * 1024 * 1024, // 16 MiB
		GrpcServerSendSize:       2 * 1024 * 1024,  // 2 MiB
		RPSLimit:                 100,
	}
	loader := confita.NewLoader(
		env.NewBackend(),
		flags.NewBackend(),
	)

	serverCtx := context.Background()

	if err := loader.Load(serverCtx, cfg); err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	var creds credentials.TransportCredentials
	if cfg.ServeTLS {
		creds, err = credentials.NewServerTLSFromFile(cfg.TLSCrtFilePath, cfg.TLSKeyFilePath)
		if err != nil {
			logger.Fatal("failed to prepare grpc transport creds", zap.Error(err))
		}
	} else {
		creds = insecure.NewCredentials()
	}

	freePort := cfg.GrpcServerBasePort
	services := make([]*grpc.Server, 0)

	signKey, err := os.ReadFile(cfg.TokenSignKeyFilePath)
	if err != nil {
		logger.Fatal("failed to read signing key", zap.Error(err))
	}

	ds, err := storage.NewDatabaseUserService(serverCtx, cfg.DatabaseConnectionString,
		time.Duration(cfg.DatabaseOperationTimeout)*time.Millisecond)
	if err != nil {
		logger.Fatal("failed to create user storage", zap.Error(err))
	}

	defer func() {
		_ = ds.Close()
	}()

	auth, err := app.NewAuthorizer(ds, signKey)
	if err != nil {
		logger.Fatal("failed to create authorizer", zap.Error(err))
	}

	authService := gsrv.NewAuthService(auth, time.Duration(cfg.DatabaseOperationTimeout)*time.Millisecond)
	storageService, _ := gsrv.NewStorageService(ds, cfg.GrpcServerSendSize)
	authFunc := gsrv.BuildAuthorizationInterceptor(auth)

	grpcServer := grpc.NewServer(grpc.Creds(creds), grpc.MaxRecvMsgSize(cfg.GrpcServerRecvSize),
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)),
		grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(gsrv.NewLimiter(int(cfg.RPSLimit)))),
		grpc.StreamInterceptor(ratelimit.StreamServerInterceptor(gsrv.NewLimiter(int(cfg.RPSLimit)))),
	)

	pb.RegisterAuthorizationServiceServer(grpcServer, authService)
	pb.RegisterStorageServer(grpcServer, storageService)

	go func(portNumber uint16) {
		addr := fmt.Sprintf("%s:%d", cfg.GrpcServerAddress, portNumber)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Fatal("failed to start grpc listener", zap.Error(err))
		}

		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("failed to serve grpc", zap.Error(err))
		}
	}(freePort)
	freePort++

	services = append(services, grpcServer)

	sCh, err := prepareShutdown(services...)
	if err != nil {
		logger.Fatal("failed to prepare shutdown", zap.Error(err))
	}

	<-sCh

	fmt.Println("Server stopped")
}

func prepareShutdown(grpcServers ...*grpc.Server) (<-chan interface{}, error) {
	shutdownSig := make(chan interface{})
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT)
	signal.Notify(signals, syscall.SIGTERM)
	signal.Notify(signals, syscall.SIGQUIT)

	go func() {
		<-signals

		for _, s := range grpcServers {
			s.GracefulStop()
		}

		close(shutdownSig)
	}()

	return shutdownSig, nil
}
