package dbhandlers

import (
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/aseptimu/url-shortener/internal/app/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type PingHandler struct {
	db     *store.Database
	logger *zap.SugaredLogger
}

func NewPingHandler(db *store.Database, logger *zap.SugaredLogger) *PingHandler {
	return &PingHandler{db, logger}
}

func (h *PingHandler) Ping(c *gin.Context) {
	utils.LogRequest(c, h.logger)

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
