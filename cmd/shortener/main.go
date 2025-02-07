package main

import (
	"log"

	"github.com/aseptimu/url-shortener/internal/app/server"
)

func main() {
	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	err := server.Run(addr)
	if err != nil {
		log.Fatalf("Server failed to start on %s: %v", addr, err)
	}
}
