package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg     *config.ConfigType
	Service service.URLShortener
}

func NewHandler(cfg *config.ConfigType, service service.URLShortener) *Handler {
	return &Handler{cfg: cfg, Service: service}
}

func (h *Handler) URLCreator(c *gin.Context) {
	defer c.Request.Body.Close()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	text := strings.TrimSpace(string(body))
	if text == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Empty body"})
		return
	}

	shortURL, err := h.Service.ShortenURL(text)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusCreated, h.cfg.BaseAddress+"/"+shortURL)
}

func (h *Handler) URLCreatorJSON(c *gin.Context) {
	defer c.Request.Body.Close()

	var req struct {
		URL string `json:"url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	shortURL, err := h.Service.ShortenURL(req.URL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"result": h.cfg.BaseAddress + "/" + shortURL})
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
