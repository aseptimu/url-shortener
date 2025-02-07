package server

import (
	"github.com/aseptimu/url-shortener/internal/app/handlers"
	"github.com/gin-gonic/gin"
)

func Run(addr string) error {
	router := gin.Default()

	router.GET("/:url", handlers.GetURL)
	router.POST("/", handlers.URLCreator)

	return router.Run(addr)
}
