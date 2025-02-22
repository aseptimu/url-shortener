package server

import (
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

	store := store.NewStore()
	service := service.NewURLService(store)
	handler := handlers.NewHandler(cfg, service)

	router.GET("/:url", handler.GetURL)
	router.POST("/", handler.URLCreator)

	return router.Run(addr)
}
