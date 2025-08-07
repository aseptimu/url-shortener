// Package dbhandlers содержит HTTP-хендлеры для проверки доступности базы данных.
package dbhandlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Pinger описывает интерфейс, который умеет «пинговать» некий ресурс (например, базу данных).
type Pinger interface {
	Ping(ctx context.Context) error
}

// PingHandler обрабатывает HTTP-запросы /ping, проверяя Pinger.
type PingHandler struct {
	db Pinger
}

// NewPingHandler создаёт новый PingHandler с переданным Pinger.
func NewPingHandler(db Pinger) *PingHandler {
	return &PingHandler{db}
}

// Ping обрабатывает GET /ping.
// Если h.db равен nil — возвращает 503 Service Unavailable.
// Иначе вызывает h.db.Ping и при ошибке возвращает 500 Internal Server Error,
// в противном случае отдаёт 200 OK.
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
