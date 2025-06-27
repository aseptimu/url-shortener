package middleware

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware_CompressesResponse(t *testing.T) {
	handler := func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	}

	r := gin.New()
	r.Use(GzipMiddleware())
	r.GET("/ping", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	r.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	gz, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	body, err := io.ReadAll(gz)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(body))
}
