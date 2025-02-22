package server

import (
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/gin-gonic/gin"
)

func Run(addr string, cfg *config.ConfigType) error {
	router := gin.Default()

	store := store.NewStore()
	service := service.NewURLService(store)
	handler := handlers.NewHandler(cfg, service)

	router.GET("/:url", handler.GetURL)
	router.POST("/", handler.URLCreator)

	return router.Run(addr)
}
