package server

import (
	"github.com/pozedorum/WB_project_3/task1/internal/service"
	"github.com/pozedorum/wbf/ginext"
)

type NotificationServer struct {
	service *service.NotificationService
}

func New(service *service.NotificationService) *NotificationServer {
	return &NotificationServer{service: service}
}

func (ns *NotificationServer) SetupRoutes(router ginext.RouterGroup) {
	notifyGroup := router.Group("/notify")
	{
		notifyGroup.POST("", ns.CreateNotification)
		notifyGroup.GET("/:id", ns.GetNotificationStatus)
		notifyGroup.DELETE("/:id", ns.DeleteNotification)
	}
	router.GET("/health", ns.HealthCheck)
}
