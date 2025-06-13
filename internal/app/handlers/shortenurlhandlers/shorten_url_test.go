package shortenurlhandlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// === mockService for Shorten, GetURL, etc. ===
type mockService struct{}

func (m *mockService) ShortenURL(_ context.Context, url string, _ string) (string, error) {
	if url == "http://example.com" {
		return "abcdef", nil
	}
	return "", errors.New("invalid URL format")
}
func (m *mockService) ShortenURLs(_ context.Context, inputs []string, _ string) (map[string]string, error) {
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
func (m *mockService) GetOriginalURL(_ context.Context, input string) (string, bool, bool) {
	if input == "abcdef" {
		return "http://example.com", true, false
	}
	return "", false, false
}
func (m *mockService) GetUserURLs(ctx context.Context, userID string) ([]service.URLRecord, error) {
	return nil, nil
}
func (m *mockService) DeleteURLs(ctx context.Context, shortURLs []string, userID string) error {
	return nil
}

func newTestHandlerShorten() *ShortenHandler {
	cfg := &config.ConfigType{BaseAddress: "http://localhost:8080"}
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	return NewShortenHandler(cfg, &mockService{}, sugar)
}

func newTestHandlerGetter() *GetURLHandler {
	cfg := &config.ConfigType{BaseAddress: "http://localhost:8080"}
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	return NewGetURLHandler(cfg, &mockService{}, sugar)
}

// === Tests for ShortenHandler and GetURLHandler ===
func TestURLCreator(t *testing.T) {
	handler := newTestHandlerShorten()
	router := gin.New()
	router.POST("/", handler.URLCreator)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
	r.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(w, r)
	res := w.Result()
	defer res.Body.Close()
	require.NotNil(t, res)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
}

func TestGetURL(t *testing.T) {
	handler := newTestHandlerGetter()
	router := gin.New()
	router.GET("/:url", handler.GetURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/abcdef", nil)
	router.ServeHTTP(w, r)
	res := w.Result()
	defer res.Body.Close()
	require.NotNil(t, res)
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, "http://example.com", res.Header.Get("Location"))
}

func TestURLCreatorJSON(t *testing.T) {
	handler := newTestHandlerShorten()
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
	bodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"result":"http://localhost:8080/abcdef"}`, string(bodyBytes))
}

func TestURLCreatorBatch(t *testing.T) {
	handler := newTestHandlerShorten()
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
	assert.JSONEq(t, `[{"correlation_id":"1","short_url":"http://localhost:8080/abcdef"}]`, string(bodyBytes))
}

// === Tests for GetUserURLs ===
type stubGetter struct {
	records []service.URLRecord
	err     error
}

func (s *stubGetter) GetOriginalURL(_ context.Context, _ string) (string, bool, bool) {
	return "", false, false
}
func (s *stubGetter) GetUserURLs(_ context.Context, _ string) ([]service.URLRecord, error) {
	return s.records, s.err
}

func newTestHandlerGetUserURLs(records []service.URLRecord, err error) *GetURLHandler {
	cfg := &config.ConfigType{BaseAddress: "http://localhost:8080"}
	logger := zap.NewNop().Sugar()
	return NewGetURLHandler(cfg, &stubGetter{records: records, err: err}, logger)
}

func TestGetUserURLs_Unauthorized_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/urls", nil)

	handler := newTestHandlerGetUserURLs(nil, nil)
	handler.GetUserURLs(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error":"Unauthorized"}`, w.Body.String())
}

func TestGetUserURLs_Unauthorized_EmptyUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/urls", nil)
	c.Set("userID", "")

	handler := newTestHandlerGetUserURLs(nil, nil)
	handler.GetUserURLs(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error":"Unauthorized"}`, w.Body.String())
}

func TestGetUserURLs_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/urls", nil)
	c.Set("userID", "user123")

	handler := newTestHandlerGetUserURLs(nil, errors.New("db failure"))
	handler.GetUserURLs(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"db failure"}`, w.Body.String())
}

func TestGetUserURLs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/urls", nil)
	c.Set("userID", "alice")

	records := []service.URLRecord{
		{ShortURL: "abc", OriginalURL: "https://go.dev"},
		{ShortURL: "xyz", OriginalURL: "https://gin-gonic.com"},
	}
	handler := newTestHandlerGetUserURLs(records, nil)
	handler.GetUserURLs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	expected := `[
		{"short_url":"http://localhost:8080/abc","original_url":"https://go.dev"},
		{"short_url":"http://localhost:8080/xyz","original_url":"https://gin-gonic.com"}
	]`
	assert.JSONEq(t, expected, w.Body.String())
}
