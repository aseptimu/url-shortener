package main

import (
	"go.uber.org/zap"
	"log"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/server"
)

func main() {
	config := config.NewConfig()
	logger, _ := zap.NewProduction()

	addr := config.ServerAddress
	log.Printf("Starting server on %s", addr)
	err := server.Run(addr, config, logger)
	if err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
}
