package server

import (
	"github.com/pozedorum/WB_project_3/task2/internal/service"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

type ShortURLServer struct {
	service service.ShortURLService
}

func New(service service.ShortURLService) *ShortURLServer {
	zlog.Logger.Info().Msg("Creating short URL server")
	return &ShortURLServer{service: service}
}

func (ss *ShortURLServer) SetupRoutes(router *ginext.RouterGroup) {
	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())

	router.POST("/shorten", ss.Shorten)
	router.GET("/s/:code", ss.Redirect)
	router.GET("/analytics/:code", ss.Analytics)
}
