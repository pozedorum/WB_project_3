package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/pozedorum/WB_project_3/task3/internal/models"
	"github.com/pozedorum/wbf/zlog"
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

func (cs *CommentService) GetCommentTree(ctx context.Context, commentID string, page int, pageSize int) (*models.CommentTreeResponse, error) {
	// Убираем пагинацию для дерева
	page = 1
	pageSize = MaxCommentsOnPage

	// Получаем все комментарии дерева
	allTreeComments, err := cs.repo.GetCommentTree(ctx, commentID)
	if err != nil {
		return nil, err
	}

	zlog.Logger.Info().
		Str("comment_id", commentID).
		Int("comments_count", len(allTreeComments)).
		Msg("comments received from database")

	if len(allTreeComments) == 0 {
		return &models.CommentTreeResponse{
			Comments:   []*models.Comment{},
			Total:      0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 0,
		}, nil
	}

	// Просто возвращаем все комментарии как плоский список
	// Не строим дерево - это нарушает порядок и логику
	flatList := make([]*models.Comment, len(allTreeComments))
	for i, comment := range allTreeComments {
		flatList[i] = &models.Comment{
			ID:        comment.ID,
			ParentID:  comment.ParentID,
			Author:    comment.Author,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Level:     cs.calculateLevel(allTreeComments, comment.ID),
		}
	}

	return &models.CommentTreeResponse{
		Comments:   flatList,
		Total:      len(flatList),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}

// calculateLevel вычисляет уровень вложенности для комментария
func (cs *CommentService) calculateLevel(comments []*models.Comment, commentID string) int {
	level := 0
	currentID := commentID

	// Находим родительские комментарии чтобы вычислить уровень
	for {
		var parentID string
		// Находим текущий комментарий
		for _, comment := range comments {
			if comment.ID == currentID {
				parentID = comment.ParentID
				break
			}
		}

		if parentID == "" {
			break
		}

		level++
		currentID = parentID
	}

	return level
}

func (cs *CommentService) GetAllComments(ctx context.Context, page int, pageSize int) (*models.CommentTreeResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > MaxCommentsOnPage {
		pageSize = DefaultPageSize
	}

	rootComments, err := cs.repo.GetRootComments(ctx)
	if err != nil {
		return nil, err
	}
	var allComments []*models.Comment
	// Преобразуем в плоский список
	for _, root := range rootComments {
		rootTree, err := cs.GetCommentTree(ctx, root.ID, page, pageSize)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, rootTree.Comments...)
	}
	totalCount := len(allComments)
	return &models.CommentTreeResponse{
		Comments:   allComments,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

// convertToFlatListWithLevels преобразует дерево в плоский список с уровнями вложенности  TODO: разобраться почему корневым коментариям не присуждается уровень
func (cs *CommentService) convertToFlatListWithLevels(comments []*models.Comment, baseLevel int) []*models.Comment {
	var result []*models.Comment

	for _, comment := range comments {
		// Создаем копию без детей, но с уровнем
		flatComment := &models.Comment{
			ID:        comment.ID,
			ParentID:  comment.ParentID,
			Author:    comment.Author,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Deleted:   comment.Deleted,
			Level:     baseLevel,
			// Children намеренно не копируем
		}
		result = append(result, flatComment)

		// Рекурсивно добавляем детей с увеличенным уровнем
		if len(comment.Children) > 0 {
			zlog.Logger.Info().Str("parrent_ID", flatComment.ID).Any("childs", comment.Children).Msg("parrent has childs")
			childComments := cs.convertToFlatListWithLevels(comment.Children, baseLevel+1)
			result = append(result, childComments...)
		}
	}

	return result
}

// Для поиска тоже добавляем преобразование
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

	// Преобразуем результаты поиска в плоский список
	flatList := cs.convertToFlatListWithLevels(comments, 0)

	return &models.SearchResponse{
		Results:    flatList,
		Query:      phrase,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (cs *CommentService) DeleteCommentTree(ctx context.Context, parentID string) error {
	return cs.repo.DeleteCommentTree(ctx, parentID)
}
