package service

import (
	"context"
	"fmt"
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

func (cs *CommentService) GetCommentTree(ctx context.Context, commentID string) (*models.CommentTreeResponse, error) {
	// Убираем пагинацию для дерева

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
			Page:       1,
			PageSize:   DefaultPageSize,
			TotalPages: 1,
		}, nil
	}

	// Строим дерево из плоского списка
	tree := cs.buildTreeFromFlatList(allTreeComments)

	// Преобразуем дерево в плоский список с DFS обходом
	flatList := cs.convertTreeToFlatListDFS(tree, 0)
	return &models.CommentTreeResponse{
		Comments:   flatList,
		Total:      len(flatList),
		Page:       1,
		PageSize:   len(flatList),
		TotalPages: 1,
	}, nil
}

func (cs *CommentService) GetAllComments(ctx context.Context) (*models.CommentTreeResponse, error) {
	// if page <= 0 {
	// 	page = 1
	// }
	// if pageSize <= 0 || pageSize > MaxCommentsOnPage {
	// 	pageSize = DefaultPageSize
	// }

	rootComments, err := cs.repo.GetRootComments(ctx)
	if err != nil {
		return nil, err
	}
	var allComments []*models.Comment
	// Преобразуем в плоский список
	for _, root := range rootComments {
		rootTree, err := cs.GetCommentTree(ctx, root.ID)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, rootTree.Comments...)
	}
	totalCount := len(allComments)

	return &models.CommentTreeResponse{
		Comments:   allComments,
		Total:      totalCount,
		Page:       1,
		PageSize:   DefaultPageSize,
		TotalPages: 1,
	}, nil
}

func (cs *CommentService) buildTreeFromFlatList(flatComments []*models.Comment) []*models.Comment {
	// Создаем карту для быстрого доступа
	commentMap := make(map[string]*models.Comment)
	var roots []*models.Comment

	// Первый проход: создаем узлы
	for _, comment := range flatComments {
		// Создаем копию с инициализированным срезом детей
		node := &models.Comment{
			ID:        comment.ID,
			ParentID:  comment.ParentID,
			Author:    comment.Author,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Deleted:   comment.Deleted,
			Children:  []*models.Comment{}, // Важно: инициализируем!
		}
		commentMap[node.ID] = node
	}

	// Второй проход: строим связи
	for _, comment := range flatComments {
		node := commentMap[comment.ID]

		if comment.ParentID != "" {
			// Если есть родитель, добавляем к нему в детей
			if parent, exists := commentMap[comment.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		} else {
			// Если родителя нет - это корневой элемент
			roots = append(roots, node)
		}
	}

	return roots
}

// convertToFlatListWithLevels преобразует дерево в плоский список с уровнями вложенности  TODO: разобраться почему корневым коментариям не присуждается уровень
func (cs *CommentService) convertTreeToFlatListDFS(tree []*models.Comment, baseLevel int) []*models.Comment {
	var result []*models.Comment

	for _, node := range tree {
		// Создаем копию узла без детей, но с уровнем
		flatNode := &models.Comment{
			ID:        node.ID,
			ParentID:  node.ParentID,
			Author:    node.Author,
			Content:   node.Content,
			CreatedAt: node.CreatedAt,
			UpdatedAt: node.UpdatedAt,
			Deleted:   node.Deleted,
			Level:     baseLevel,
			// Children намеренно не копируем
		}

		// Добавляем текущий узел
		result = append(result, flatNode)

		// Рекурсивно добавляем детей (DFS)
		if len(node.Children) > 0 {
			childNodes := cs.convertTreeToFlatListDFS(node.Children, baseLevel+1)
			result = append(result, childNodes...)
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

	allComments, totalCount, err := cs.repo.SearchComments(ctx, phrase, page, pageSize)
	if err != nil {
		return nil, err
	}
	//tree := cs.buildTreeFromFlatList(allComments)
	// Преобразуем результаты поиска в плоский список
	//flatList := cs.convertTreeToFlatListDFS(tree, 0)

	return &models.SearchResponse{
		Results:    allComments,
		Query:      phrase,
		Total:      totalCount,
		Page:       1,
		PageSize:   totalCount,
		TotalPages: 1,
	}, nil
}

func (cs *CommentService) DeleteCommentTree(ctx context.Context, parentID string) error {
	return cs.repo.DeleteCommentTree(ctx, parentID)
}
