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

	logger.Infow("Initializing server", "address", addr)

	router := gin.New()

	logger.Debug("Setting up middleware")
	router.Use(middleware.MiddlewareLogger(logger))
	router.Use(middleware.GzipMiddleware())

	var sourceStore service.Store = db

	if cfg.DSN != "" {
		logger.Debugw("Database mode enabled, initializing tables")
		db.CreateTables(logger)
	} else {
		logger.Debugw("File storage mode enabled", "storagePath", cfg.FileStoragePath)
		sourceStore = store.NewFileStore(cfg.FileStoragePath)
	}

	urlService := service.NewURLService(sourceStore)
	handler := handlers.NewHandler(cfg, urlService, db)

	router.GET("/:url", handler.GetURL)
	router.GET("/ping", handler.Ping)
	router.POST("/", handler.URLCreator)
	router.POST("/api/shorten", handler.URLCreatorJSON)
	router.POST("/api/shorten/batch", handler.URLCreatorBatch)

	logger.Debugw("Starting server", "address", addr)
	err := router.Run(addr)
	if err != nil {
		logger.Errorw("Server failed to start", "error", err)
	}
	return err
}
