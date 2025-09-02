package server

import (
	"strconv"
	"time"

	"github.com/pozedorum/WB_project_3/task3/internal/models"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

// Работает
func (cs *CommentServer) PostNewComment(c *ginext.Context) {
	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for create new comment")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request: " + err.Error()})
		return
	}
	if request.Author == "" || request.Content == "" {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Author and content are required"})
		return
	}
	zlog.Logger.Info().
		Str("parent_id", request.ParentID).
		Str("author", request.Author).
		Str("content", request.Content).
		Msg("Creating new comment")

	comment, err := cs.service.PostNewComment(c.Request.Context(), request)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("parent_id", request.ParentID).
			Str("author", request.Author).
			Msg("Failed to create comment")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().
		Str("comment_id", comment.ID).
		Str("author", comment.Author).
		Msg("Comment created successfully")

	c.JSON(models.StatusAccepted, comment)
}

// Работает
func (cs *CommentServer) GetCommentTree(c *ginext.Context) {
	commentID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := cs.service.GetCommentTree(c.Request.Context(), commentID, page, pageSize)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("comment_id", commentID).
			Int("page", page).
			Int("page_size", pageSize).
			Msg("Failed to get comment tree")
		c.JSON(models.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().
		Str("comment_id", commentID).
		Int("page", page).
		Int("page_size", pageSize).
		Int("total", result.Total).
		Msg("Comment tree retrieved successfully")

	c.JSON(models.StatusOK, result)
}

// Работает
func (cs *CommentServer) DeleteCommentTree(c *ginext.Context) {
	parrentID := c.Param("id")
	if parrentID == "" {
		zlog.Logger.Error().Msg("Empty comment ID in delete request")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Comment ID is required"})
		return
	}

	zlog.Logger.Info().
		Str("comment_id", parrentID).
		Msg("Deleting comment tree")

	err := cs.service.DeleteCommentTree(c.Request.Context(), parrentID)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("comment_id", parrentID).
			Msg("Failed to delete comment tree")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().
		Str("comment_id", parrentID).
		Msg("Comment tree deleted successfully")

	c.JSON(models.StatusOK, ginext.H{
		"message":    "Comment tree deleted successfully",
		"comment_id": parrentID,
	})
}

func (cs *CommentServer) SearchComments(c *ginext.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if query == "" {
		zlog.Logger.Error().Msg("Empty search query")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Search query is required"})
		return
	}

	zlog.Logger.Info().
		Str("query", query).
		Int("page", page).
		Int("page_size", pageSize).
		Msg("Searching comments")

	result, err := cs.service.SearchComments(c.Request.Context(), query, page, pageSize)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("query", query).
			Msg("Failed to search comments")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().
		Str("query", query).
		Int("total_results", result.Total).
		Int("page", page).
		Int("page_size", pageSize).
		Msg("Search completed successfully")

	c.JSON(models.StatusOK, result)
}

// Работает
func (cs *CommentServer) GetAllComments(c *ginext.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	zlog.Logger.Info().
		Int("page", page).
		Int("page_size", pageSize).
		Msg("Getting all comments")

	result, err := cs.service.GetAllComments(c.Request.Context(), page, pageSize)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Int("page", page).
			Int("page_size", pageSize).
			Msg("Failed to get all comments")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().
		Int("page", page).
		Int("page_size", pageSize).
		Int("total_comments", result.Total).
		Msg("All comments retrieved successfully")

	c.JSON(models.StatusOK, result)
}

func (cs *CommentServer) HealthCheck(c *ginext.Context) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "comment-tree-service",
	}

	zlog.Logger.Debug().Msg("Health check completed")
	c.JSON(models.StatusOK, response)
}
