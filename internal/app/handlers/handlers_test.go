package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct{}

func (m *mockService) ShortenURL(url string) (string, error) {
	if url == "http://example.com" {
		return "abcdef", nil
	}
	return "", errors.New("invalid URL format")
}

func (m *mockService) GetOriginalURL(input string) (string, bool) {
	if input == "abcdef" {
		return "http://example.com", true
	}
	return "", false
}

func TestURLCreator(t *testing.T) {
	cfg := &config.ConfigType{BaseAddress: "http://localhost:8080"}
	handler := NewHandler(cfg, &mockService{})

	router := gin.Default()
	router.POST("/", handler.URLCreator)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
	r.Header.Set("Content-Type", "text/plain") // Добавляем заголовок

	router.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	require.NotNil(t, res)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
}

func TestGetURL(t *testing.T) {
	cfg := &config.ConfigType{BaseAddress: "http://localhost:8080"}
	handler := NewHandler(cfg, &mockService{})

	router := gin.Default()
	router.GET("/:url", handler.GetURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/abcdef", nil)

	router.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	require.NotNil(t, res)
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, "http://example.com", res.Header.Get("Location")) // Проверяем редирект
}
