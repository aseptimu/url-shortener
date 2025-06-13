package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockService struct{}

func (m *mockService) ShortenURL(_ context.Context, url string) (string, error) {
	if url == "http://example.com" {
		return "abcdef", nil
	}
	return "", errors.New("invalid URL format")
}
func (m *mockService) ShortenURLs(_ context.Context, inputs []string) (map[string]string, error) {
	shortened := make(map[string]string, len(inputs))
	for _, input := range inputs {
		if input == "http://example.com" {
			shortened["abcdef"] = input
		} else {
			return nil, errors.New("invalid URL format")
		}
	}
	return shortened, nil
}

func (m *mockService) GetOriginalURL(_ context.Context, input string) (string, bool) {
	if input == "abcdef" {
		return "http://example.com", true
	}
	return "", false
}

func newTestHandler() *ShortenHandler {
	cfg := &config.ConfigType{BaseAddress: "http://localhost:8080"}

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	return NewShortenHandler(cfg, &mockService{}, sugar)
}

func TestURLCreator(t *testing.T) {
	handler := newTestHandler()

	router := gin.New()
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
	handler := newTestHandler()

	router := gin.New()
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

func TestURLCreatorJSON(t *testing.T) {
	handler := newTestHandler()

	router := gin.New()
	router.POST("/api/shorten", handler.URLCreatorJSON)

	w := httptest.NewRecorder()
	jsonBody := `{"url": "http://example.com"}`
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(jsonBody))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	require.NotNil(t, res)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	expectedResponse := `{"result":"http://localhost:8080/abcdef"}`
	bodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.JSONEq(t, expectedResponse, string(bodyBytes))
}

func TestURLCreatorBatch(t *testing.T) {
	handler := newTestHandler()

	router := gin.New()
	router.POST("/batch", handler.URLCreatorBatch)

	w := httptest.NewRecorder()
	batchReq := `[{"correlation_id":"1","original_url":"http://example.com"}]`
	r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(batchReq))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	require.NotNil(t, res)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	bodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	expected := `[{"correlation_id":"1","short_url":"http://localhost:8080/abcdef"}]`
	assert.JSONEq(t, expected, string(bodyBytes))
}
