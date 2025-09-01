package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/pozedorum/WB_project_3/task3/internal/models"
)

const (
	MaxCommentsOnPage = 15
	DefaultPageSize   = 10
)

type CommentService struct {
	repo Repository
}

func NewCommentService(repo Repository) *CommentService {
	return &CommentService{repo: repo}
}

func (cs *CommentService) PostNewComment(ctx context.Context, req models.CreateCommentRequest) (*models.Comment, error) {
	newCom := &models.Comment{
		ID:        uuid.New().String(),
		ParentID:  req.ParentID,
		Author:    req.Author,
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := cs.repo.CreateComment(ctx, newCom)
	if err != nil {
		return nil, err
	}
	return newCom, nil
}

func (cs *CommentService) GetCommentTree(ctx context.Context, parentID string, page int, pageSize int) (*models.CommentTreeResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > MaxCommentsOnPage {
		pageSize = DefaultPageSize
	}

	comments, totalCount, err := cs.repo.GetCommentsByParentID(ctx, parentID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var trees []*models.Comment
	if parentID == "" {
		trees = comments
	} else {
		trees = cs.buildTrees(comments)
	}

	return &models.CommentTreeResponse{
		Comments:   trees,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (cs *CommentService) GetAllComments(ctx context.Context, page int, pageSize int) (*models.CommentTreeResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > MaxCommentsOnPage {
		pageSize = DefaultPageSize
	}

	allComments, totalCount, err := cs.repo.GetCommentsByParentID(ctx, "", page, pageSize)
	if err != nil {
		return nil, err
	}
	roots := cs.buildTrees(allComments)

	return &models.CommentTreeResponse{
		Comments:   roots,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (cs *CommentService) DeleteCommentTree(ctx context.Context, parentID string) error {
	return cs.repo.DeleteCommentTree(ctx, parentID)
}

func (cs *CommentService) SearchComments(ctx context.Context, phrase string, page, pageSize int) (*models.SearchResponse, error) {
	if phrase == "" {
		return nil, fmt.Errorf("empty search phrase")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > MaxCommentsOnPage {
		pageSize = DefaultPageSize
	}

	comments, totalCount, err := cs.repo.SearchComments(ctx, phrase, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &models.SearchResponse{
		Results:    comments,
		Query:      phrase,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (cs *CommentService) buildTrees(comments []*models.Comment) []*models.Comment {
	commentMap := make(map[string]*models.Comment)
	var roots []*models.Comment

	// Сначала создаем мапу для быстрого доступа
	for _, comment := range comments {
		commentMap[comment.ID] = comment
	}

	// Затем строим иерархию
	for _, comment := range comments {
		if comment.ParentID == "" {
			roots = append(roots, comment)
		} else {
			if parent, exists := commentMap[comment.ParentID]; exists {
				parent.Children = append(parent.Children, comment)
			}
		}
	}

	return roots
}
