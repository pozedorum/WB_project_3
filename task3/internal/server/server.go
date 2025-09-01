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

	// API роуты
	router.POST("/comments", cs.PostNewComment)
	router.DELETE("/comments/:id", cs.DeleteCommentTree)
	router.GET("/comments", cs.GetCommentTree)
	router.GET("/comments/all", cs.GetAllComments)
	router.GET("/comments/search", cs.SearchComments)
	router.GET("/health", cs.HealthCheck)
}
