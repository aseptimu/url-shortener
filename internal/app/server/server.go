// Package server настраивает маршруты, middleware и запускает HTTP-сервер.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers/dbhandlers"
	"github.com/aseptimu/url-shortener/internal/app/handlers/shortenurlhandlers"
	"github.com/aseptimu/url-shortener/internal/app/middleware"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/aseptimu/url-shortener/internal/app/utils"
	"github.com/aseptimu/url-shortener/internal/app/workers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Run инициализирует маршруты, подключает middleware и запускает сервер на адресе addr.
func Run(addr string, cfg *config.ConfigType, logger *zap.SugaredLogger) error {
	gin.SetMode(gin.ReleaseMode)

	logger.Infow("Initializing server", "address", addr)

	router := gin.New()

	logger.Debug("Setting up middleware")
	router.Use(middleware.MiddlewareLogger(logger))
	router.Use(middleware.GzipMiddleware())

	secretKey := cfg.SecretKey
	if cfg.SecretKey == "" {
		logger.Warn("SecretKey not provided, generating a random one (will reset on each restart)")
		cfg.SecretKey = utils.GenerateRandomSecretKey()
	}
	router.Use(middleware.AuthMiddleware(secretKey, logger))

	var db *store.Database
	var sourceStore service.Store = db
	logger.Debugw("Connecting to database", "DB config", cfg.DSN)
	if cfg.DSN != "" {
		if err := store.MigrateDB(cfg.DSN, logger); err != nil {
			logger.Fatalf("Database migration failed: %v", err)
		}

		db = store.NewDB(cfg.DSN, logger)
		logger.Debugw("Database mode enabled, initializing tables")
		sourceStore = db
	} else {
		logger.Debugw("File storage mode enabled", "storagePath", cfg.FileStoragePath)
		sourceStore = store.NewFileStore(cfg.FileStoragePath)
	}

	urlService := service.NewURLService(sourceStore)
	urlGetService := service.NewGetURLService(sourceStore)
	urlDelete := service.NewURLDeleter(sourceStore)

	getURLHandler := shortenurlhandlers.NewGetURLHandler(cfg, urlGetService, logger)
	shortenHandler := shortenurlhandlers.NewShortenHandler(cfg, urlService, logger)
	deleteURLHandler := shortenurlhandlers.NewDeleteURLHandler(cfg, urlDelete, logger)
	pingHandler := dbhandlers.NewPingHandler(db)

	router.GET("/:url", getURLHandler.GetURL)
	router.GET("/ping", pingHandler.Ping)
	router.POST("/", shortenHandler.URLCreator)
	router.POST("/api/shorten", shortenHandler.URLCreatorJSON)
	router.POST("/api/shorten/batch", shortenHandler.URLCreatorBatch)
	router.GET("/api/user/urls", getURLHandler.GetUserURLs)
	router.DELETE("/api/user/urls", deleteURLHandler.DeleteUserURLs)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	workers.StartDeleteWorkerPool(ctx, 5, urlDelete, logger)

	srv := &http.Server{
		Addr:    addr,
		Handler: router.Handler(),
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		logger.Infow("Shutting down server", "signal", "signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Errorw("Error shutting down server", "error", err)
		}
	}()

	if *cfg.EnableHTTPS {
		certPEM, keyPEM, err := utils.GenerateSelfSignedCert()
		if err != nil {
			logger.Fatalf("Не удалось сгенерировать сертификат: %v", err)
		}
		tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			logger.Fatalf("Ошибка создания X509KeyPair: %v", err)
		}

		srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{tlsCert}}

		logger.Infow("Запуск HTTPS сервера", "addr", addr)
		err = srv.ListenAndServeTLS("", "")
		wg.Wait()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	logger.Infow("Запуск HTTP сервера", "addr", addr)
	err := srv.ListenAndServe()
	wg.Wait()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
