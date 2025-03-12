package handlers

import (
	"encoding/json"
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

type Handler struct {
	cfg     *config.ConfigType
	Service service.URLShortener
	db      *store.Database
	logger  *zap.SugaredLogger
}

func NewHandler(cfg *config.ConfigType, service service.URLShortener, db *store.Database, logger *zap.SugaredLogger) *Handler {
	return &Handler{cfg: cfg, Service: service, db: db, logger: logger}
}

func (h *Handler) Ping(c *gin.Context) {
	h.logRequest(c)

	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Server doesn't use database"})
		return
	}

	err := h.db.Ping()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func (h *Handler) URLCreator(c *gin.Context) {
	h.logRequest(c)

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

	shortURL, err := h.Service.ShortenURL(text.String())
	if err != nil && !errors.Is(err, service.ErrConflict) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")
	if errors.Is(err, service.ErrConflict) {
		c.String(http.StatusConflict, h.cfg.BaseAddress+"/"+shortURL)
	} else {
		c.String(http.StatusCreated, h.cfg.BaseAddress+"/"+shortURL)
	}
}

func (h *Handler) URLCreatorJSON(c *gin.Context) {
	h.logRequest(c)

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

	shortURL, err := h.Service.ShortenURL(req.URL)
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

func (h *Handler) URLCreatorBatch(c *gin.Context) {
	h.logRequest(c)

	var requestURLs []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&requestURLs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	responseURLs := make([]URLResponse, len(requestURLs))

	conflict := false

	for i, requestURL := range requestURLs {
		shortURL, err := h.Service.ShortenURL(requestURL.OriginalURL)
		if err != nil {
			if errors.Is(err, service.ErrConflict) {
				conflict = true
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
				return
			}
		}

		responseURLs[i] = URLResponse{
			CorrelationID: requestURL.CorrelationID,
			ShortURL:      h.cfg.BaseAddress + "/" + shortURL,
		}
	}

	c.Header("Content-Type", "application/json")
	if conflict {
		c.JSON(http.StatusConflict, responseURLs)
	} else {
		c.JSON(http.StatusCreated, responseURLs)
	}
}

func (h *Handler) GetURL(c *gin.Context) {
	h.logRequest(c)
	key := c.Param("url")
	originalURL, exists := h.Service.GetOriginalURL(key)
	if !exists {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Header("Location", originalURL)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusTemporaryRedirect, originalURL)
}

func (h *Handler) logRequest(c *gin.Context) {
	h.logger.Debugw("Endpoint called",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"remote_addr", c.ClientIP(),
	)
}
