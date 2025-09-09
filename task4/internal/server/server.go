package server

import (
	"github.com/pozedorum/WB_project_3/task4/internal/service"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

type ImageProcessServer struct {
	service *service.ImageProcessService
}

func New(service *service.ImageProcessService) *ImageProcessServer {
	zlog.Logger.Info().Msg("Creating Image Processing Server")
	return &ImageProcessServer{service: service}
}

func (ips *ImageProcessServer) SetupRoutes(router *ginext.RouterGroup) {
	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())
	router.Use(CORSMiddleware())
	// // Статические файлы
	router.Static("/static", "./internal/frontend/static")

	// Главная страница
	// Главная страница
	router.GET("/", ips.ServeFrontend)

	router.POST("/upload", ips.UploadNewImage)
	router.GET("/image/:id", ips.GetProcessedImage)
	router.DELETE("/image/:id", ips.DeleteImage)
}

func CORSMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
