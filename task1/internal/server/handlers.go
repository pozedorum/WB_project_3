package server

import (
	"net/http"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/wbf/ginext"
)

func (ns *NotificationServer) CreateNotification(c *ginext.Context) {
	var req models.CreateNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	n, err := ns.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	resp := models.NotificationResponse{
		ID:      n.ID,
		Status:  n.Status,
		SendAt:  n.SendAt,
		Message: n.Message,
	}
	c.JSON(http.StatusAccepted, resp)
}

func (ns *NotificationServer) GetNotificationStatus(c *ginext.Context) {
	id := c.Param("id")

	n, err := ns.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if n == nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "notification not found"})
		return
	}

	resp := models.NotificationResponse{
		ID:      n.ID,
		Status:  n.Status,
		SendAt:  n.SendAt,
		Message: n.Message,
	}

	c.JSON(http.StatusOK, resp)
}

func (ns *NotificationServer) DeleteNotification(c *ginext.Context) {
	id := c.Param("id")

	if err := ns.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"status": "canceled"})
}

func (ns *NotificationServer) HealthCheck(c *ginext.Context) {
	c.JSON(http.StatusOK, ginext.H{
		"status":  "ok",
		"service": "notification-server",
	})
}
