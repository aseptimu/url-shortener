// Package shortenurlhandlers содержит HTTP-хендлеры для операций с короткими URL.
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

// DeleteTask представляет задачу пакетного удаления списка коротких URL
// для конкретного пользователя.
type DeleteTask struct {
	URLs   []string
	UserID string
}

// DeleteTaskCh — буферизированный канал для передачи задач удаления.
// Работает совместно с фоновым воркером.
var DeleteTaskCh = make(chan DeleteTask, 100)

// DeleteURLHandler обрабатывает HTTP-запросы на удаление
// собственных сокращённых ссылок пользователя.
type DeleteURLHandler struct {
	cfg     *config.ConfigType
	Service service.URLDeleter
	logger  *zap.SugaredLogger
}

// NewDeleteURLHandler создаёт новый экземпляр DeleteURLHandler,
// принимая на вход конфиг приложения, сервис удаления URL и логгер.
func NewDeleteURLHandler(cfg *config.ConfigType, service service.URLDeleter, logger *zap.SugaredLogger) *DeleteURLHandler {
	return &DeleteURLHandler{cfg: cfg, Service: service, logger: logger}
}

// DeleteUserURLs обрабатывает DELETE /api/user/urls.
// Читает JSON-массив коротких ссылок из тела запроса,
// формирует задачу DeleteTask и отправляет её в канал DeleteTaskCh.
// Требует наличия валидного userID в контексте (JWT в куках).
// В случае успеха возвращает HTTP 202 Accepted.
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
