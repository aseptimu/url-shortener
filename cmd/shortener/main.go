package main

import (
	"github.com/aseptimu/url-shortener/internal/app/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleRoute)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
