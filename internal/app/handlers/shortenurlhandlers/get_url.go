// Package shortenurlhandlers содержит HTTP-хендлеры для операций с короткими URL.
package shortenurlhandlers

import (
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

// GetURLHandler обрабатывает перенаправление на оригинальный URL и выдачу списка URL пользователя.
type GetURLHandler struct {
	cfg     *config.ConfigType
	Service service.URLGetter
	logger  *zap.SugaredLogger
}

// NewGetURLHandler создаёт новый экземпляр GetURLHandler.
func NewGetURLHandler(cfg *config.ConfigType, service service.URLGetter, logger *zap.SugaredLogger) *GetURLHandler {
	return &GetURLHandler{cfg: cfg, Service: service, logger: logger}
}

// GetURL перенаправляет клиента на оригинальный URL, если он существует и не удалён.
func (h *GetURLHandler) GetURL(c *gin.Context) {
	utils.LogRequest(c, h.logger)

	key := c.Param("url")
	originalURL, exists, deleted := h.Service.GetOriginalURL(c.Request.Context(), key)
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}
	if deleted {
		c.AbortWithStatusJSON(http.StatusGone, gin.H{"error": "URL is deleted"})
		return
	}

	c.Header("Location", originalURL)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusTemporaryRedirect, originalURL)
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

	records, err := h.Service.GetUserURLs(c.Request.Context(), userIDStr)
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
