package handlers

import (
	"encoding/json"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
)

type Handler struct {
	cfg     *config.ConfigType
	Service service.URLShortener
	db      *store.Database
}

func NewHandler(cfg *config.ConfigType, service service.URLShortener, db *store.Database) *Handler {
	return &Handler{cfg: cfg, Service: service, db: db}
}

func (h *Handler) Ping(c *gin.Context) {
	defer c.Request.Body.Close()

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
	defer c.Request.Body.Close()

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

	shortURL, err, isConflict := h.Service.ShortenURL(text.String())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")
	if isConflict {
		c.String(http.StatusConflict, h.cfg.BaseAddress+"/"+shortURL)
	} else {
		c.String(http.StatusCreated, h.cfg.BaseAddress+"/"+shortURL)
	}
}

func (h *Handler) URLCreatorJSON(c *gin.Context) {
	defer c.Request.Body.Close()

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

	shortURL, err, isConflict := h.Service.ShortenURL(req.URL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := struct {
		Result string `json:"result"`
	}{
		Result: h.cfg.BaseAddress + "/" + shortURL,
	}

	if isConflict {
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
	defer c.Request.Body.Close()

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
		shortURL, err, isConflict := h.Service.ShortenURL(requestURL.OriginalURL)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
			return
		}
		if isConflict {
			conflict = true
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
