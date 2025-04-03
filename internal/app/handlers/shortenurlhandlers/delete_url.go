package shortenurlhandlers

import (
	"encoding/json"
	"net/http"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeleteTask struct {
	URLs   []string
	UserID string
}

var DeleteTaskCh = make(chan DeleteTask, 100)

type DeleteURLHandler struct {
	cfg     *config.ConfigType
	Service service.URLDeleter
	logger  *zap.SugaredLogger
}

func NewDeleteURLHandler(cfg *config.ConfigType, service service.URLDeleter, logger *zap.SugaredLogger) *DeleteURLHandler {
	return &DeleteURLHandler{cfg: cfg, Service: service, logger: logger}
}

func (h *DeleteURLHandler) DeleteUserURLs(c *gin.Context) {
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

	var urls []string
	if err := json.NewDecoder(c.Request.Body).Decode(&urls); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	task := DeleteTask{
		URLs:   urls,
		UserID: userIDStr,
	}

	DeleteTaskCh <- task

	c.Status(http.StatusAccepted)
}
