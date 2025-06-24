// Package middleware содержит Gin-middleware для упаковки и распаковки
// HTTP-сообщений в формате gzip.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipWriter оборачивает gin.ResponseWriter и gzip.Writer,
// чтобы записи в ResponseWriter автоматически сжимались.
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write сжимает переданные данные в формате gzip и записывает их
// во внутренний gzip.Writer.
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// GzipMiddleware возвращает Gin-middleware, который:
//  1. при входящем запросе с заголовком Content-Encoding: gzip
//     распаковывает тело запроса;
//  2. при наличии Accept-Encoding: gzip в заголовках запроса
//     сжимает исходящий ответ в формате gzip.
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Gzip content"})
				return
			}
			defer reader.Close()
			c.Request.Body = io.NopCloser(reader)
		}

		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		gzWriter := gzip.NewWriter(c.Writer)
		defer gzWriter.Close()

		c.Writer = &gzipWriter{ResponseWriter: c.Writer, writer: gzWriter}
		c.Header("Content-Encoding", "gzip")

		c.Next()
	}
}
