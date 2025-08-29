package models

import (
	"errors"
	"time"

	"github.com/pozedorum/wbf/retry"
)

type Comment struct {
	ID        string     `json:"id"`
	ParentID  string     `json:"parent_id,omitempty"`
	Author    string     `json:"author"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Deleted   bool       `json:"-"`
	Children  []*Comment `json:"children,omitempty"`
}

type CreateCommentRequest struct {
	ParentID string `json:"parent_id"`
	Author   string `json:"author"`
	Content  string `json:"content"`
}

type CommentTreeResponse struct {
	CommentRoot *Comment `json:"comments"`
	Total       int      `json:"total"`
	Page        int      `json:"page"`
	PageSize    int      `json:"page_size"`
	TotalPages  int      `json:"total_pages"`
}

type CommentsResponse struct {
	Comments []*Comment `json:"comments"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type SearchRequest struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

var (
	StandardStrategy = retry.Strategy{Attempts: 3, Delay: time.Second}
	ConsumerStrategy = retry.Strategy{Attempts: 5, Delay: 2 * time.Second}
)

var (
	ErrShortURLNotFound   = errors.New("short URL not found")
	ErrDuplicateShortCode = errors.New("duplicate short code")
)

const (
	StatusOK                  = 200
	StatusAccepted            = 202
	StatusFound               = 302
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatisConflict            = 409
	StatusInternalServerError = 500
)
