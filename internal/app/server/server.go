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

func Run(addr string, cfg *config.ConfigType, logger *zap.SugaredLogger) error {
	gin.SetMode(gin.ReleaseMode)

	logger.Infow("Initializing server", "address", addr)

	router := gin.New()

	logger.Debug("Setting up middleware")
	router.Use(middleware.MiddlewareLogger(logger))
	router.Use(middleware.GzipMiddleware())

	var db *store.Database
	var sourceStore service.Store = db
	logger.Debugw("Connecting to database", "DB config", cfg.DSN)
	if cfg.DSN != "" {
		db = store.NewDB(cfg.DSN, logger)
		logger.Debugw("Database mode enabled, initializing tables")
		db.CreateTables(logger)
		sourceStore = db
	} else {
		logger.Debugw("File storage mode enabled", "storagePath", cfg.FileStoragePath)
		sourceStore = store.NewFileStore(cfg.FileStoragePath)
	}

	urlService := service.NewURLService(sourceStore)
	shortenHandler := handlers.NewShortenHandler(cfg, urlService, logger)
	pingHandler := handlers.NewPingHandler(db)

	router.GET("/:url", shortenHandler.GetURL)
	router.GET("/ping", pingHandler.Ping)
	router.POST("/", shortenHandler.URLCreator)
	router.POST("/api/shorten", shortenHandler.URLCreatorJSON)
	router.POST("/api/shorten/batch", shortenHandler.URLCreatorBatch)

	logger.Debugw("Starting server", "address", addr)
	err := router.Run(addr)
	if err != nil {
		logger.Errorw("Server failed to start", "error", err)
	}
	return err
}
