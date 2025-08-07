// Package server настраивает маршруты, middleware и запускает HTTP-сервер.
package http

import (
	"context"
	"crypto/tls"
	"errors"
	http2 "github.com/aseptimu/url-shortener/internal/app/handlers/http"
	"github.com/aseptimu/url-shortener/internal/app/middleware"
	"github.com/aseptimu/url-shortener/internal/app/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	srv    *http.Server
	logger *zap.SugaredLogger
}

func NewServer(addr string, secretKey string, logger *zap.SugaredLogger, h http2.Handlers) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	logger.Debug("Setting up middleware")
	r.Use(middleware.MiddlewareLogger(logger), middleware.GzipMiddleware(), middleware.AuthMiddleware(secretKey, logger))
	h.RegisterRoutes(r)

	return &Server{
		srv:    &http.Server{Addr: addr, Handler: r},
		logger: logger,
	}
}

// Run инициализирует маршруты, подключает middleware и запускает сервер на адресе addr.
func (s *Server) Run(ctx context.Context, enableHTTPS bool) error {
	s.logger.Infow("Initializing server", "address", s.srv.Addr)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		s.logger.Infow("Shutting down server", "signal", "signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Errorw("Error shutting down server", "error", err)
		}
	}()

	if enableHTTPS {
		certPEM, keyPEM, err := utils.GenerateSelfSignedCert()
		if err != nil {
			s.logger.Fatalf("Не удалось сгенерировать сертификат: %v", err)
		}
		tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			s.logger.Fatalf("Ошибка создания X509KeyPair: %v", err)
		}

		s.srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{tlsCert}}

		s.logger.Infow("Запуск HTTPS сервера", "addr", s.srv.Addr)
		err = s.srv.ListenAndServeTLS("", "")
		wg.Wait()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	} else {
		s.logger.Infow("Запуск HTTP сервера", "addr", s.srv.Addr)
		err := s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		wg.Wait()
	}

	return nil
}
