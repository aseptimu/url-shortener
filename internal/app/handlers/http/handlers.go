package http

import (
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers/http/dbhandlers"
	"github.com/aseptimu/url-shortener/internal/app/handlers/http/shortenurlhandlers"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handlers interface {
	RegisterRoutes(r *gin.Engine)
}

type handlersImpl struct {
	cfg          *config.ConfigType
	urlSvc       service.URLShortener
	urlGetSvc    shortenurlhandlers.URLGetter
	urlDeleteSvc service.URLDeleter
	pinger       dbhandlers.Pinger
	logger       *zap.SugaredLogger
}

func New(
	cfg *config.ConfigType,
	urlSvc service.URLShortener,
	urlGetSvc shortenurlhandlers.URLGetter,
	urlDeleteSvc service.URLDeleter,
	pinger dbhandlers.Pinger,
	logger *zap.SugaredLogger,
) Handlers {
	return &handlersImpl{
		cfg:          cfg,
		urlSvc:       urlSvc,
		urlGetSvc:    urlGetSvc,
		urlDeleteSvc: urlDeleteSvc,
		pinger:       pinger,
		logger:       logger,
	}
}

func (h *handlersImpl) RegisterRoutes(r *gin.Engine) {
	r.GET("/:url", shortenurlhandlers.NewGetURLHandler(h.cfg, h.urlGetSvc, h.logger).GetURL)
	r.GET("/ping", dbhandlers.NewPingHandler(h.pinger).Ping)
	r.POST("/", shortenurlhandlers.NewShortenHandler(h.cfg, h.urlSvc, h.logger).URLCreator)
	r.POST("/api/shorten", shortenurlhandlers.NewShortenHandler(h.cfg, h.urlSvc, h.logger).URLCreatorJSON)
	r.POST("/api/shorten/batch", shortenurlhandlers.NewShortenHandler(h.cfg, h.urlSvc, h.logger).URLCreatorBatch)
	r.GET("/api/user/urls", shortenurlhandlers.NewGetURLHandler(h.cfg, h.urlGetSvc, h.logger).GetUserURLs)
	r.DELETE("/api/user/urls", shortenurlhandlers.NewDeleteURLHandler(h.cfg, h.urlDeleteSvc, h.logger).DeleteUserURLs)
	r.GET("/api/internal/stats", shortenurlhandlers.NewGetURLHandler(h.cfg, h.urlGetSvc, h.logger).GetStats)
}
