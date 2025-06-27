package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestMiddlewareLogger(t *testing.T) {
	core, obs := observer.New(zap.InfoLevel)
	logger := zap.New(core).Sugar()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MiddlewareLogger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "hello")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := obs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	msg := entries[0].Message
	if !strings.Contains(msg, "Request URI /test Method GET") {
		t.Errorf("log message missing request info: %s", msg)
	}
	if !strings.Contains(msg, "Response status 200 Response size 5") {
		t.Errorf("log message missing response info: %s", msg)
	}
}
