// Package middleware содержит Gin-middleware для логирования HTTP-запросов.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		gin.ResponseWriter
		responseData *responseData
	}
)

// Write записывает данные в базовый ResponseWriter и увеличивает счётчик размера ответа.
func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.responseData.size += size
	return size, err
}

// WriteHeader записывает код статуса в базовый ResponseWriter и сохраняет его в responseData.
func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.status = statusCode
}

// MiddlewareLogger возвращает Gin-middleware, которое логирует каждый запрос.
// В логах выводятся URI, HTTP-метод, время обработки, статус ответа и размер ответа.
func MiddlewareLogger(sugar *zap.SugaredLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		responseData := &responseData{
			size:   0,
			status: 0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: ctx.Writer,
			responseData:   responseData,
		}

		ctx.Writer = &lw

		now := time.Now()
		ctx.Next()
		duration := time.Since(now)

		sugar.Infoln(
			"Request",
			"URI", ctx.Request.URL.Path,
			"Method", ctx.Request.Method,
			"Duration", duration,
			"Response status", responseData.status,
			"Response size", responseData.size,
		)
	}
}
