package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task3/internal/models"
)

// Repository интерфейс для работы с данными комментариев
type Repository interface {
	// Создание комментария
	CreateComment(ctx context.Context, comment *models.Comment) error
	// Получение комментария по ID
	GetCommentByID(ctx context.Context, id string) (*models.Comment, error)

	// Получение корневых
	GetRootComments(ctx context.Context) ([]*models.Comment, error)
	// Получение дерева комментариев
	GetCommentTree(ctx context.Context, commentID string) ([]*models.Comment, error)
	// Получение всех комментариев (для построения дерева)
	GetAllComments(ctx context.Context) ([]*models.Comment, error)
	// Удаление комментария и всех его потомков
	DeleteCommentTree(ctx context.Context, id string) error
	// Поиск комментариев по тексту (полнотекстовый поиск)
	SearchComments(ctx context.Context, query string, page, pageSize int) ([]*models.Comment, int, error)
	// Обновление комментария (если потребуется редактирование)
	UpdateComment(ctx context.Context, comment *models.Comment) error
}
