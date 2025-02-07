package main

import (
	"flag"
	"log"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/server"
)

func main() {
	flag.Parse()
	addr := config.Config.ServerAddress
	log.Printf("Starting server on %s", addr)
	err := server.Run(addr)
	if err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
}
