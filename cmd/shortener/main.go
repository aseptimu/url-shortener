package main

import (
	"fmt"
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
