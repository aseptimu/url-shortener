package http

import (
	"context"
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/handlers/http/dbhandlers"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeDB struct {
	err error
}

func (f *fakeDB) Ping(_ context.Context) error {
	return f.err
}

func TestPing_NoDB(t *testing.T) {
	handler := dbhandlers.NewPingHandler(nil)
	router := gin.New()
	router.GET("/ping", handler.Ping)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	body, _ := io.ReadAll(res.Body)
	assert.JSONEq(t, `{"error":"Server doesn't use database"}`, string(body))
}

func TestPing_DBFails(t *testing.T) {
	handler := dbhandlers.NewPingHandler(&fakeDB{err: errors.New("fail ping")})
	router := gin.New()
	router.GET("/ping", handler.Ping)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	body, _ := io.ReadAll(res.Body)
	assert.JSONEq(t, `{"error":"fail ping"}`, string(body))
}

func TestPing_OK(t *testing.T) {
	handler := dbhandlers.NewPingHandler(&fakeDB{err: nil})
	router := gin.New()
	router.GET("/ping", handler.Ping)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, _ := io.ReadAll(res.Body)
	assert.Empty(t, string(body))
}
