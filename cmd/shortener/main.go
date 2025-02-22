package main

import (
	"log"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/server"
)

func main() {
	config := config.NewConfig()

	addr := config.ServerAddress
	log.Printf("Starting server on %s", addr)
	err := server.Run(addr, config)
	if err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
}
