package main

import (
	"flag"
	"log"
	"os"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/server"
)

func main() {
	flag.Parse()

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		config.Config.ServerAddress = serverAddress
	}

	if baseAddress := os.Getenv("BASE_URL"); baseAddress != "" {
		config.Config.BaseAddress = baseAddress
	}

	addr := config.Config.ServerAddress
	log.Printf("Starting server on %s", addr)
	err := server.Run(addr)
	if err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
}
