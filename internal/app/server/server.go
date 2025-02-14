package server

import (
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers"
	"github.com/gin-gonic/gin"
)

func Run(addr string, cfg *config.ConfigType) error {
	router := gin.Default()

	handler := handlers.NewHandler(cfg)

	router.GET("/:url", handler.GetURL)
	router.POST("/", handler.URLCreator)

	return router.Run(addr)
}
