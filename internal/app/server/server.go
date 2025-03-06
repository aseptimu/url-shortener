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

func Run(addr string, cfg *config.ConfigType, db *store.Database, logger *zap.SugaredLogger) error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(middleware.MiddlewareLogger(logger))
	router.Use(middleware.GzipMiddleware())

	var sourceStore service.Store = db
	if cfg.DSN != "" {
		db.CreateTables(logger)
	} else {
		sourceStore = store.NewFileStore(cfg.FileStoragePath)
	}

	urlService := service.NewURLService(sourceStore)
	handler := handlers.NewHandler(cfg, urlService, db)

	router.GET("/:url", handler.GetURL)
	router.GET("/ping", handler.Ping)
	router.POST("/", handler.URLCreator)
	router.POST("/api/shorten", handler.URLCreatorJSON)
	router.POST("/api/shorten/batch", handler.URLCreatorBatch)

	return router.Run(addr)
}
