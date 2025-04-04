package utils

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LogRequest(c *gin.Context, logger *zap.SugaredLogger) {
	logger.Debugw("Endpoint called",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"remote_addr", c.ClientIP(),
	)
}
