package main

import (
	"github.com/aseptimu/url-shortener/internal/app/config"
	"go.uber.org/zap"
	"log"
	"net/http"

	"github.com/aseptimu/url-shortener/internal/app/server"
	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	appCfg := config.NewConfig()

	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.DisableCaller = true
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

	addr := appCfg.ServerAddress
	log.Printf("Starting server on %s", addr)

	if err := server.Run(addr, appCfg, sugar); err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
}
