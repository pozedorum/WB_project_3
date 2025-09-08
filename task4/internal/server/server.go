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

func (is *ImageProcessServer) SetupRoutes(router *ginext.RouterGroup) {
	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())

	// // Статические файлы
	router.Static("/static", "./internal/frontend/templates")

	// Главная страница
	router.GET("/", func(c *ginext.Context) {
		c.File("./internal/frontend/templates/index.html")
	})

	router.POST("/upload", is.UploadNewImage)
	router.GET("/image/:id", is.GetProcessedImage)
	router.DELETE("/image/:id", is.DeleteImage)
}
