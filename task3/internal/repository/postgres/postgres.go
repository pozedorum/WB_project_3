package postgres

import (
	"context"
	"time"

	"github.com/pozedorum/WB_project_3/task3/internal/models"
	"github.com/pozedorum/wbf/dbpg"
	"github.com/pozedorum/wbf/zlog"
)

type CommentRepository struct {
	db *dbpg.DB
}

func NewCommentRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*CommentRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewCommentRepository(db), nil
}

func NewCommentRepository(db *dbpg.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (nr *CommentRepository) Close() {
	if err := nr.db.Master.Close(); err != nil {
		zlog.Logger.Panic().Msg("Database failed to close")
	}
	for _, slave := range nr.db.Slaves {
		if slave != nil {
			if err := slave.Close(); err != nil {
				zlog.Logger.Panic().Msg("Slave database failed to close")
			}
		}
	}
	zlog.Logger.Info().Msg("PostgreSQL connections closed")
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (id, parent_id, author, content, created_at, updated_at, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var parentID interface{}
	if comment.ParentID == "" {
		parentID = nil
	} else {
		parentID = comment.ParentID
	}
	_, err := r.db.ExecWithRetry(ctx, models.StandardStrategy, query,
		comment.ID,
		parentID,
		comment.Author,
		comment.Content,
		comment.CreatedAt,
		comment.UpdatedAt,
		comment.Deleted,
	)

	return err
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	query := `
		SELECT id, COALESCE(parent_id, '') as parent_id, author, content, created_at, updated_at, deleted
		FROM comments 
		WHERE id = $1 AND deleted = false
	`

	var comment models.Comment

	rows, err := r.db.QueryWithRetry(ctx, models.StandardStrategy, query, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zlog.Logger.Panic().Msg("failed to close sql rows")
		}
	}()
	if !rows.Next() {
		return nil, nil
	}

	err = rows.Scan(
		&comment.ID,
		&comment.ParentID,
		&comment.Author,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.Deleted,
	)
	if err != nil {
		return nil, err
	}

	// if parentID != nil {
	// 	comment.ParentID = *parentID
	// } else {
	// 	comment.ParentID = ""
	// }

	return &comment, nil
}

func (r *CommentRepository) GetRootComments(ctx context.Context) ([]*models.Comment, error) {

	query := `SELECT id, COALESCE(parent_id, '') as parent_id, author, content, created_at, updated_at, deleted 
			  FROM comments 
			  WHERE parent_id IS NULL AND deleted = false`

	rows, err := r.db.QueryWithRetry(ctx, models.StandardStrategy, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zlog.Logger.Panic().Msg("failed to close sql rows")
		}
	}()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment

		err = rows.Scan(
			&comment.ID,
			&comment.ParentID,
			&comment.Author,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.Deleted,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *CommentRepository) GetCommentTree(ctx context.Context, commentID string) ([]*models.Comment, error) {
	query := `
		WITH RECURSIVE comment_tree AS (
			-- Начинаем с запрошенного комментария (включаем его самого)
			SELECT id, parent_id, author, content, created_at, updated_at, deleted
			FROM comments 
			WHERE id = $1 AND deleted = false
			
			UNION ALL
			
			-- Ищем ВСЕХ потомков (детей, внуков и т.д.)
			SELECT c.id, c.parent_id, c.author, c.content, c.created_at, c.updated_at, c.deleted
			FROM comments c
			INNER JOIN comment_tree ct ON c.parent_id = ct.id  -- Ищем детей текущего узла
			WHERE c.deleted = false
		)
		SELECT id, parent_id, author, content
		FROM comment_tree
		ORDER BY created_at
	`
	zlog.Logger.Info().Str("comment_id", commentID).Msg("Getting comment tree from DB")
	rows, err := r.db.QueryWithRetry(ctx, models.StandardStrategy, query, commentID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zlog.Logger.Panic().Msg("failed to close sql rows")
		}
	}()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		var parentID *string // Для nullable parent_id

		err = rows.Scan(
			&comment.ID,
			&parentID, // Сканируем как указатель
			&comment.Author,
			&comment.Content,
			// &comment.CreatedAt,
			// &comment.UpdatedAt,
			// &comment.Deleted,
		)
		if err != nil {
			return nil, err
		}

		// Обрабатываем nullable parent_id
		if parentID != nil {
			comment.ParentID = *parentID
		} else {
			comment.ParentID = ""
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	zlog.Logger.Info().Int("count", len(comments)).Msg("Comments found")
	for _, comment := range comments {
		zlog.Logger.Info().
			Str("id", comment.ID).
			Str("parent_id", comment.ParentID).
			Msg("Comment in tree")
	}
	return comments, nil
}

func (r *CommentRepository) GetAllComments(ctx context.Context) ([]*models.Comment, error) {
	query := `
		SELECT id, COALESCE(parent_id, '') as parent_id, author, content, created_at, updated_at, deleted
		FROM comments 
		WHERE deleted = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryWithRetry(ctx, models.StandardStrategy, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zlog.Logger.Panic().Msg("failed to close sql rows")
		}
	}()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment

		err = rows.Scan(
			&comment.ID,
			&comment.ParentID,
			&comment.Author,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.Deleted,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *CommentRepository) DeleteCommentTree(ctx context.Context, id string) error {
	// Используем рекурсивный CTE для удаления всего поддерева
	query := `
		WITH RECURSIVE comment_tree AS (
			SELECT id FROM comments WHERE id = $1
			UNION ALL
			SELECT c.id FROM comments c
			INNER JOIN comment_tree ct ON c.parent_id = ct.id
		)
		UPDATE comments 
		SET deleted = true, updated_at = $2
		WHERE id IN (SELECT id FROM comment_tree)
	`

	_, err := r.db.ExecWithRetry(ctx, models.StandardStrategy, query, id, time.Now())
	return err
}

func (r *CommentRepository) SearchComments(ctx context.Context, query string, page, pageSize int) ([]*models.Comment, int, error) {
	// ILIKE - функция регистронезависимого поиска
	searchQuery := `
		SELECT id, COALESCE(parent_id, '') as parent_id, author, content, created_at, updated_at, deleted
		FROM comments 
		WHERE deleted = false AND content ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*) 
		FROM comments 
		WHERE deleted = false AND content ILIKE $1
	`

	// Добавляем % для поиска по подстроке
	searchPattern := "%" + query + "%"

	// Получаем общее количество
	var totalCount int
	rows, err := r.db.QueryWithRetry(ctx, models.StandardStrategy, countQuery, searchPattern)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zlog.Logger.Panic().Msg("failed to close sql rows")
		}
	}()

	if rows.Next() {
		err = rows.Scan(&totalCount)
		if err != nil {
			return nil, 0, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	// Получаем данные с пагинацией
	offset := (page - 1) * pageSize
	rows, err = r.db.QueryWithRetry(ctx, models.StandardStrategy, searchQuery, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			zlog.Logger.Panic().Msg("failed to close sql rows")
		}
	}()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment

		err = rows.Scan(
			&comment.ID,
			&comment.ParentID,
			&comment.Author,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.Deleted,
		)
		if err != nil {
			return nil, 0, err
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return comments, totalCount, nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, comment *models.Comment) error {
	query := `
		UPDATE comments 
		SET content = $1, updated_at = $2
		WHERE id = $3 AND deleted = false
	`

	_, err := r.db.ExecWithRetry(ctx, models.StandardStrategy, query,
		comment.Content,
		time.Now(),
		comment.ID,
	)

	return err
}

// Проверка имплементации интерфейса репозитория
//var _ service.Repository = (*CommentRepository)(nil)
