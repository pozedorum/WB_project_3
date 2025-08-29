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
	}

	err := cs.repo.CreateComment(ctx, newCom)
	if err != nil {
		return nil, err
	}
	return newCom, nil
}

func (cs *CommentService) GetCommentTree(ctx context.Context, parentID string, page int, pageSize int) (*models.CommentTreeResponse, error) {
	if parentID == "" {
		return nil, fmt.Errorf("empty root id")
	}
	if pageSize <= 0 || pageSize > MaxCommentsOnPage {
		pageSize = MaxCommentsOnPage
	}

	if page <= 0 {
		page = 1
	}

	comments, totalCount, err := cs.repo.GetCommentsByParentID(ctx, parentID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Для корневых комментариев (parentID = "") можем построить дерево
	//var tree []*models.Comment
	// if parentID == "" {
	// 	// Получаем ВСЕ комментарии для построения полного дерева
	// 	allComments, err := cs.repo.GetAllComments(ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	tree = cs.buildTree(allComments)
	// } else {
	// 	tree = comments
	// }
	roots := cs.buildTree(comments)
	if len(roots) > 1 {
		return nil, fmt.Errorf("multiple root comments: %d", len(roots))
	}
	return &models.CommentTreeResponse{
		CommentRoot: roots[0], //tree
		Total:       totalCount,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (cs *CommentService) GetAllComments(ctx context.Context, page int, pageSize int) ([]*models.CommentTreeResponse, error) {

	allComments, err := cs.repo.GetAllComments(ctx)
	if err != nil {
		return nil, err
	}
	tree := cs.buildTree(allComments)
	res := make([]*models.CommentTreeResponse, len(tree))
	for _, root := range tree {
		newResp := &models.CommentTreeResponse{
			CommentRoot: root,
		}
		res = append(res, newResp)
	}
}

func (cs *CommentService) DeleteCommentTree(ctx context.Context, parentID string) error {
	err := cs.repo.DeleteCommentTree(ctx, parentID)
	if err != nil {
		return err
	}
	return nil
}

func (cs *CommentService) SearchComments(ctx context.Context, phrase string, page, pageSize int) ([]*models.Comment, int, error)

func (cs *CommentService) buildTree(comments []*models.Comment) []*models.Comment {
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
