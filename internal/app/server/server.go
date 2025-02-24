package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/store"
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

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.responseData.size += size
	return size, err
}

func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.status = statusCode
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func gzipMiddleware() gin.HandlerFunc {
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

func middlewareLogger(sugar *zap.SugaredLogger) gin.HandlerFunc {
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

func Run(addr string, cfg *config.ConfigType) error {
	gin.SetMode(gin.ReleaseMode)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	router := gin.New()

	sugar := logger.Sugar()
	router.Use(middlewareLogger(sugar))
	router.Use(gzipMiddleware())

	store := store.NewFileStore(cfg.FileStoragePath)
	service := service.NewURLService(store)
	handler := handlers.NewHandler(cfg, service)

	router.GET("/:url", handler.GetURL)
	router.POST("/", handler.URLCreator)
	router.POST("/api/shorten", handler.URLCreatorJSON)

	return router.Run(addr)
}
