package server

import (
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers"
	"github.com/aseptimu/url-shortener/internal/app/middleware"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(addr string, cfg *config.ConfigType, logger *zap.Logger) error {
	gin.SetMode(gin.ReleaseMode)

	defer logger.Sync()

	router := gin.New()

	sugar := logger.Sugar()
	router.Use(middleware.MiddlewareLogger(sugar))
	router.Use(middleware.GzipMiddleware())

	store := store.NewFileStore(cfg.FileStoragePath)
	service := service.NewURLService(store)
	handler := handlers.NewHandler(cfg, service)

	router.GET("/:url", handler.GetURL)
	router.POST("/", handler.URLCreator)
	router.POST("/api/shorten", handler.URLCreatorJSON)

	return router.Run(addr)
}
