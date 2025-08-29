package server

import (
	"time"

	"github.com/pozedorum/WB_project_3/task3/internal/models"
	"github.com/pozedorum/wbf/ginext"
)

func (cs *CommentServer) PostNewComment(c *ginext.Context) {

}

func (cs *CommentServer) GetCommentTree(c *ginext.Context) {

}

func (cs *CommentServer) DeleteCommentTree(c *ginext.Context) {

}

func (ss *CommentServer) HealthCheck(c *ginext.Context) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "url-shortener",
	}
	c.JSON(models.StatusOK, response)
}
