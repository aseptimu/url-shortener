package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type PingHandler struct {
	db Pinger
}

func NewPingHandler(db Pinger) *PingHandler {
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
