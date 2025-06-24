// Package utils содержит вспомогательные функции.
package utils

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LogRequest выполняет DEBUG-логирование входящего HTTP-запроса.
// В лог сохраняются метод запроса, полный путь (FullPath) и IP клиента.
func LogRequest(c *gin.Context, logger *zap.SugaredLogger) {
	logger.Debugw("Endpoint called",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"remote_addr", c.ClientIP(),
	)
}
