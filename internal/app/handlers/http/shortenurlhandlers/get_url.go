// Package shortenurlhandlers содержит HTTP-хендлеры для операций с короткими URL.
package shortenurlhandlers

import (
	"context"
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
)

// URLGetter предоставляет методы получения URL для клиентского кода.
type URLGetter interface {
	GetOriginalURL(ctx context.Context, input string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]service.URLDTO, error)
	GetStats(ctx context.Context) (service.StatsDTO, error)
}

// Stats хранит данные о количестве пользователей и сохраненных url
type Stats struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}

// URLRecord хранит данные одной записи сокращённого URL.
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// GetURLHandler обрабатывает перенаправление на оригинальный URL и выдачу списка URL пользователя.
type GetURLHandler struct {
	cfg     *config.ConfigType
	service URLGetter
	logger  *zap.SugaredLogger
}

// NewGetURLHandler создаёт новый экземпляр GetURLHandler.
func NewGetURLHandler(cfg *config.ConfigType, service URLGetter, logger *zap.SugaredLogger) *GetURLHandler {
	return &GetURLHandler{cfg: cfg, service: service, logger: logger}
}

// GetURL перенаправляет клиента на оригинальный URL, если он существует и не удалён.
func (h *GetURLHandler) GetURL(c *gin.Context) {
	utils.LogRequest(c, h.logger)

	key := c.Param("url")
	originalURL, err := h.service.GetOriginalURL(c.Request.Context(), key)
	switch {
	case errors.Is(err, service.ErrURLNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrURLDeleted):
		c.AbortWithStatusJSON(http.StatusGone, gin.H{"error": err.Error()})
	}

	c.Header("Location", originalURL)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusTemporaryRedirect, originalURL)
}

// GetStats возвращает кол-во url и пользователей
func (h *GetURLHandler) GetStats(c *gin.Context) {
	utils.LogRequest(c, h.logger)

	if h.cfg.TrustedSubnet == "" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	clientIP := net.ParseIP(c.GetHeader("X-Real-IP"))
	_, trustedNet, err := net.ParseCIDR(h.cfg.TrustedSubnet)
	if err != nil {
		h.logger.Errorw("Invalid CIDR in config.TrustedSubnet", "value", h.cfg.TrustedSubnet, "error", err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if clientIP == nil || !trustedNet.Contains(clientIP) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Errorw("Failed to get stats", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetUserURLs возвращает все короткие URL, созданные текущим пользователем.
func (h *GetURLHandler) GetUserURLs(c *gin.Context) {
	utils.LogRequest(c, h.logger)

	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	records, err := h.service.GetUserURLs(c.Request.Context(), userIDStr)
	if err != nil {
		h.logger.Errorw("Failed to get user URLs", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(records) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	var resp []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	for _, rec := range records {
		resp = append(resp, struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}{
			ShortURL:    h.cfg.BaseAddress + "/" + rec.ShortURL,
			OriginalURL: rec.OriginalURL,
		})
	}

	c.JSON(http.StatusOK, resp)
}
