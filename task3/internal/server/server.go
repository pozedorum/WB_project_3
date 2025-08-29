package server

import (
	"github.com/pozedorum/WB_project_3/task3/internal/service"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

type CommentServer struct {
	service *service.CommentService
}

func New(service *service.CommentService) *CommentServer {
	zlog.Logger.Info().Msg("Creating comment server")
	return &CommentServer{service: service}
}

func (cs *CommentServer) SetupRoutes(router *ginext.RouterGroup) {
	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())

	// Фронтенд роуты
	// router.GET("/", ss.IndexPage)
	// router.GET("/result", ss.ResultPage)
	// router.GET("/analytics-page", ss.AnalyticsPage)

	// API роуты
	router.POST("/comments", cs.PostNewComment)
	router.GET("/comments", cs.GetCommentTree)
	router.DELETE("/comments/:id", cs.DeleteCommentTree)
	router.GET("/health", cs.HealthCheck)
}
