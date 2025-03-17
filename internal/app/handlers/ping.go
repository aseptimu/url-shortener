package handlers

import (
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PingHandler struct {
	db *store.Database
}

func NewPingHandler(db *store.Database) *PingHandler {
	return &PingHandler{db}
}

func (h *PingHandler) Ping(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Server doesn't use database"})
		return
	}

	err := h.db.Ping(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
