package shorten_url_handlers

import (
	"encoding/json"
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/utils"
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
	utils.LogRequest(c, h.logger)

	var userIDStr string
	if uid, exists := c.Get("userID"); exists {
		if str, ok := uid.(string); ok {
			userIDStr = str
		}
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

	shortURL, err := h.Service.ShortenURL(c.Request.Context(), text.String(), userIDStr)
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
	utils.LogRequest(c, h.logger)

	var userIDStr string
	if uid, exists := c.Get("userID"); exists {
		if str, ok := uid.(string); ok {
			userIDStr = str
		}
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

	shortURL, err := h.Service.ShortenURL(c.Request.Context(), req.URL, userIDStr)
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
	utils.LogRequest(c, h.logger)

	var userIDStr string
	if uid, exists := c.Get("userID"); exists {
		if str, ok := uid.(string); ok {
			userIDStr = str
		}
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
	shortenedURLs, err := h.Service.ShortenURLs(c.Request.Context(), inputURLs, userIDStr)
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
