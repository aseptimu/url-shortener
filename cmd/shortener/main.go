package main

import (
	"context"
	"fmt"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers/grpc/proto"
	http2 "github.com/aseptimu/url-shortener/internal/app/handlers/http"
	"github.com/aseptimu/url-shortener/internal/app/handlers/http/dbhandlers"
	"github.com/aseptimu/url-shortener/internal/app/middleware"
	grpcServer "github.com/aseptimu/url-shortener/internal/app/server/grpc"
	httpServer "github.com/aseptimu/url-shortener/internal/app/server/http"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/aseptimu/url-shortener/internal/app/workers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	_ "net/http/pprof"
)

func main() {
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.DisableCaller = false
	zapCfg.DisableStacktrace = true
	zapCfg.Sampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("can't build zap logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	go func() {
		sugar.Errorw("Error running pprof server", http.ListenAndServe("localhost:6060", nil))
	}()
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)

	appCfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	var storeSvc service.Store
	var pinger dbhandlers.Pinger
	sugar.Debugw("Connecting to database", "DB config", appCfg.DSN)
	if appCfg.DSN != "" {
		if err := store.MigrateDB(appCfg.DSN, sugar); err != nil {
			sugar.Fatalf("Database migration failed: %v", err)
		}
		db := store.NewDB(appCfg.DSN, sugar)
		sugar.Debugw("Database mode enabled, initializing tables")
		storeSvc, pinger = db, db
	} else {
		sugar.Debugw("File storage mode enabled", "storagePath", appCfg.FileStoragePath)
		storeSvc = store.NewFileStore(appCfg.FileStoragePath)
	}

	urlSvc := service.NewURLService(storeSvc)
	urlGet := service.NewGetURLService(storeSvc)
	urlDel := service.NewURLDeleter(storeSvc)

	h := http2.New(
		appCfg,
		urlSvc,
		urlGet,
		urlDel,
		pinger,
		sugar,
	)

	addr := appCfg.ServerAddress
	sugar.Infow("Starting server on", "address: ", addr)

	srv := httpServer.NewServer(addr, appCfg.SecretKey, sugar, h)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	workers.StartDeleteWorkerPool(ctx, 5, urlDel, sugar)

	grpcImpl := grpcServer.NewServer(
		appCfg,
		urlSvc,
		urlGet,
		urlDel,
		pinger)

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.AuthInterceptor(appCfg.SecretKey, sugar)),
	)

	proto.RegisterURLShortenerServer(grpcSrv, grpcImpl)

	sugar.Infow("Starting gRPC server on", "address: ", appCfg.GRPCServerAddress)
	lis, err := net.Listen("tcp", appCfg.GRPCServerAddress)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			sugar.Errorf("gRPC server stopped with error: %v", err)
		}
	}()

	if err := srv.Run(ctx, *appCfg.EnableHTTPS); err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
	<-ctx.Done()
	grpcSrv.GracefulStop()
}
