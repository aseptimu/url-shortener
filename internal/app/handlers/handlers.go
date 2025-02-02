package handlers

import (
	"github.com/aseptimu/url-shortener/internal/app/service"
	"io"
	"net/http"
	"strings"
)

func URLCreator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	text := strings.TrimSpace(string(body))

	shortURL, err := service.ShortenURL(text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte("http://localhost:8080/" + shortURL))
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/")
	if key == "" {
		http.Error(w, "URL missed", http.StatusBadRequest)
		return
	}

	originalURL, exists := service.GetOriginalURL(key)
	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(originalURL))
}

func HandleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		GetURL(w, r)
	} else if r.Method == http.MethodPost {
		URLCreator(w, r)
	} else {
		http.Error(w, "Only GET and POST methods allowed", http.StatusMethodNotAllowed)
	}
}
