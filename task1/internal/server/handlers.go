package server

import (
	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

func (ns *NotificationServer) CreateNotification(c *ginext.Context) {
	var req models.CreateNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for create notification")
		c.JSON(models.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().
		Str("user_id", req.UserID).
		Str("channel", req.Channel).
		Time("send_at", req.SendAt).
		Msg("Creating notification")

	n, err := ns.service.Create(c.Request.Context(), &req)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("user_id", req.UserID).
			Str("channel", req.Channel).
			Msg("Failed to create notification")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	resp := models.NotificationResponse{
		ID:      n.ID,
		Status:  n.Status,
		SendAt:  n.SendAt,
		Message: n.Message,
		Channel: n.Channel,
	}

	zlog.Logger.Info().
		Str("notification_id", n.ID).
		Str("status", n.Status).
		Msg("Notification created successfully")

	c.JSON(models.StatusAccepted, resp)
}

func (ns *NotificationServer) GetNotificationStatus(c *ginext.Context) {
	id := c.Param("id")

	zlog.Logger.Info().Str("notification_id", id).Msg("Getting notification status")

	n, err := ns.service.GetByID(c.Request.Context(), id)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", id).Msg("Failed to get notification")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if n == nil {
		zlog.Logger.Warn().Str("notification_id", id).Msg("Notification not found")
		c.JSON(models.StatusNotFound, ginext.H{"error": "notification not found"})
		return
	}

	resp := models.NotificationResponse{
		ID:      n.ID,
		Status:  n.Status,
		SendAt:  n.SendAt,
		Message: n.Message,
		Channel: n.Channel,
	}

	zlog.Logger.Info().
		Str("notification_id", id).
		Str("status", n.Status).
		Msg("Notification status retrieved")

	c.JSON(models.StatusOK, resp)
}

func (ns *NotificationServer) DeleteNotification(c *ginext.Context) {
	id := c.Param("id")

	zlog.Logger.Info().Str("notification_id", id).Msg("Deleting notification")

	if err := ns.service.Delete(c.Request.Context(), id); err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", id).Msg("Failed to delete notification")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Str("notification_id", id).Msg("Notification deleted successfully")
	c.JSON(models.StatusOK, ginext.H{"status": "canceled"})
}

func (ns *NotificationServer) HealthCheck(c *ginext.Context) {
	zlog.Logger.Debug().Msg("Health check requested")
	c.JSON(models.StatusOK, ginext.H{
		"status":  "healthy",
		"service": "notification-server",
	})
}
