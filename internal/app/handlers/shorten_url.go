package handlers

import (
	"encoding/json"
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

type ShortenHandler struct {
	cfg     *config.ConfigType
	Service service.URLShortener
	logger  *zap.SugaredLogger
}

func NewShortenHandler(cfg *config.ConfigType, service service.URLShortener, logger *zap.SugaredLogger) *ShortenHandler {
	return &ShortenHandler{cfg: cfg, Service: service, logger: logger}
}

func (h *ShortenHandler) URLCreator(c *gin.Context) {
	h.logRequest(c)

	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	text, err := url.Parse(string(body))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Empty body"})
		return
	}

	shortURL, err := h.Service.ShortenURL(c.Request.Context(), text.String(), userID.(string))
	if err != nil && !errors.Is(err, service.ErrConflict) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")
	if errors.Is(err, service.ErrConflict) {
		c.String(http.StatusConflict, h.cfg.BaseAddress+"/"+shortURL)
	} else {
		c.String(http.StatusCreated, h.cfg.BaseAddress+"/"+shortURL)
	}
}

func (h *ShortenHandler) URLCreatorJSON(c *gin.Context) {
	h.logRequest(c)

	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	shortURL, err := h.Service.ShortenURL(c.Request.Context(), req.URL, userID.(string))
	if err != nil && !errors.Is(err, service.ErrConflict) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := struct {
		Result string `json:"result"`
	}{
		Result: h.cfg.BaseAddress + "/" + shortURL,
	}

	if errors.Is(err, service.ErrConflict) {
		c.JSON(http.StatusConflict, resp)
	} else {
		c.JSON(http.StatusCreated, resp)
	}
}

type URLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (h *ShortenHandler) URLCreatorBatch(c *gin.Context) {
	h.logRequest(c)

	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var requestURLs []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&requestURLs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	inputURLs := make([]string, len(requestURLs))
	for i, req := range requestURLs {
		inputURLs[i] = req.OriginalURL
	}

	// Функция ShortenURLs возвращает map[shortURL]originalURL
	shortenedURLs, err := h.Service.ShortenURLs(c.Request.Context(), inputURLs, userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URLs"})
		return
	}

	// Преобразуем в map[originalURL]shortURL
	shortenedURLsMap := make(map[string]string, len(shortenedURLs))
	for short, orig := range shortenedURLs {
		shortenedURLsMap[orig] = short
	}

	responseURLs := make([]URLResponse, len(requestURLs))
	for i, req := range requestURLs {
		shortURL, ok := shortenedURLsMap[req.OriginalURL]
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Mismatch in shortened URLs"})
			return
		}
		responseURLs[i] = URLResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      h.cfg.BaseAddress + "/" + shortURL,
		}
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusCreated, responseURLs)
}

func (h *ShortenHandler) GetURL(c *gin.Context) {
	h.logRequest(c)
	key := c.Param("url")
	originalURL, exists := h.Service.GetOriginalURL(c.Request.Context(), key)
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Header("Location", originalURL)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusTemporaryRedirect, originalURL)
}

func (h *ShortenHandler) GetUserURLs(c *gin.Context) {
	h.logRequest(c)

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

func (h *ShortenHandler) logRequest(c *gin.Context) {
	h.logger.Debugw("Endpoint called",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"remote_addr", c.ClientIP(),
	)
}
