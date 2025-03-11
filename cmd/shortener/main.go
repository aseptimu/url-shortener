package main

import (
	"github.com/aseptimu/url-shortener/internal/app/store"
	"go.uber.org/zap"
	"log"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/server"
)

func main() {
	config := config.NewConfig()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()

	var db *store.Database
	sugar.Debugw("Connecting to database", "DB config", config.DSN)
	if config.DSN != "" {
		db = store.NewDB(config.DSN, sugar)
	}

	addr := config.ServerAddress
	log.Printf("Starting server on %s", addr)
	err := server.Run(addr, config, db, sugar)
	if err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
}
