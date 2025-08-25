package server

import (
	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/service"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

type NotificationServer struct {
	service *service.NotificationService
}

func New(service *service.NotificationService) *NotificationServer {
	zlog.Logger.Info().Msg("Creating notification server")
	return &NotificationServer{service: service}
}

func (ns *NotificationServer) SetupRoutes(router *ginext.RouterGroup) {
	router.GET("/", func(c *ginext.Context) {
		zlog.Logger.Debug().Msg("Serving index page")
		c.HTML(models.StatusOK, "index.html", nil)
	})

	notifyGroup := router.Group("/notify")
	{
		notifyGroup.POST("", ns.CreateNotification)
		notifyGroup.GET("/:id", ns.GetNotificationStatus)
		notifyGroup.DELETE("/:id", ns.DeleteNotification)
	}
	router.GET("/health", ns.HealthCheck)

	zlog.Logger.Info().Msg("Notification server routes configured")
}
