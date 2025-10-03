package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task7/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task7/internal/models"
	"github.com/wb-go/wbf/zlog"
)

type WarehouseService struct {
	repo interfaces.Repository
}

func New(repo interfaces.Repository) *WarehouseService {
	logger.LogService(func() {
		zlog.Logger.Info().Msg("Creating new WarehouseService instance")
	})
	return &WarehouseService{repo: repo}
}

func (servs *WarehouseService) CreateItem(ctx context.Context, req *models.ItemRequest, username string, role models.UserRole) (*models.Item, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Str("username", username).
			Str("role", string(role)).
			Str("item name", req.Name).
			Int64("price", req.Price).
			Msg("Creating new sale")
	})

	if err := validateItemRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Str("item name", req.Name).
				Int64("price", req.Price).
				Msg("Sale request validation failed")
		})
		return nil, err
	}
	if role == models.RoleViewer || role == models.RoleManager {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(models.ErrNotEnoughRights).
				Str("role", string(role)).
				Str("item name", req.Name).
				Int64("price", req.Price).
				Msg("not enough rights")
		})
		return nil, models.ErrNotEnoughRights
	}
	newItem := &models.Item{
		Name:      req.Name,
		Price:     req.Price,
		CreatedBy: username,
	}

	if err := servs.repo.Create(ctx, newItem); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Str("item name", req.Name).
				Int64("price", req.Price).
				Msg("create sale in database failed")
		})
		return nil, err
	}
	return newItem, nil
}

func (servs *WarehouseService) GetItemByID(ctx context.Context, id int64) (*models.Item, error) {
	var (
		res *models.Item
		err error
	)
	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Getting item by id")
	})
	if id <= 0 {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(models.ErrInvalidID).
				Int64("id", id).
				Msg("id is negative or null")
		})
		return nil, models.ErrInvalidID
	}

	if res, err = servs.repo.FindByID(ctx, id); err != nil || res == nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Int64("id", id).
				Msg("failed to found item in database")
		})
	}

	return res, nil
}

func (servs *WarehouseService) GetAllItems(ctx context.Context) ([]models.Item, error) {
	var (
		res []models.Item
		err error
	)
	logger.LogService(func() {
		zlog.Logger.Info().
			Msg("Getting all items")
	})
	if res, err = servs.repo.FindAll(ctx); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Msg("failed to get items from database")
		})
		return nil, err
	}
	return res, nil
}

func (servs *WarehouseService) UpdateItem(ctx context.Context, id int64, req *models.ItemRequest, username string, role models.UserRole) (*models.Item, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Str("username", username).
			Str("role", string(role)).
			Str("item name", req.Name).
			Int64("price", req.Price).
			Msg("Update sale")
	})

	if err := validateItemRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Str("item name", req.Name).
				Int64("price", req.Price).
				Msg("Sale request validation failed")
		})
		return nil, err
	}
	if role == models.RoleViewer {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(models.ErrNotEnoughRights).
				Str("role", string(role)).
				Str("item name", req.Name).
				Int64("price", req.Price).
				Msg("not enough rights")
		})
		return nil, models.ErrNotEnoughRights
	}
	newItem := &models.Item{
		Name:      req.Name,
		Price:     req.Price,
		CreatedBy: username,
	}

	if err := servs.repo.Update(ctx, id, newItem); err != nil {
		if err == models.ErrNoRows {
			logger.LogRepository(func() {
				zlog.Logger.Warn().
					Int64("id", id).
					Msg("No rows affected - item not found")
			})
			return nil, models.ErrNoRows // Запись не найдена
		}
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Str("item name", req.Name).
				Int64("price", req.Price).
				Msg("create sale in database failed")
		})
		return nil, err
	}
	return newItem, nil
}

func (servs *WarehouseService) DeleteItem(ctx context.Context, id int64, username string, role models.UserRole) error {
	var (
		res *models.Item
		err error
	)
	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Getting item by id")
	})

	if role == models.RoleViewer || role == models.RoleManager {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(models.ErrNotEnoughRights).
				Str("role", string(role)).
				Str("item name", username).
				Msg("not enough rights")
		})
		return models.ErrNotEnoughRights
	}

	if id <= 0 {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Int64("id", id).
				Msg("id is negative or null")
		})
		return models.ErrInvalidID
	}

	if err = servs.repo.Delete(ctx, id); err != nil || res == nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Int64("id", id).
				Msg("failed to delete item from database")
		})
	}

	return nil
}

func (servs *WarehouseService) GetItemHistory(ctx context.Context, itemID int64, role models.UserRole) ([]models.HistoryRecord, error) {
	var (
		res []models.HistoryRecord
		err error
	)

	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", itemID).
			Msg("Getting item history by id")
	})

	if itemID <= 0 {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(models.ErrInvalidID).
				Int64("id", itemID).
				Msg("id is negative or null")
		})
		return nil, models.ErrInvalidID
	}

	if res, err = servs.repo.GetItemHistory(ctx, itemID); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Int64("itemID", itemID).
				Msg("failed to get item history from database")
		})
	}
	return res, nil
}

func (servs *WarehouseService) GetAllHistory(ctx context.Context, filters map[string]interface{}, role models.UserRole) ([]models.HistoryRecord, error) {
	var (
		res []models.HistoryRecord
		err error
	)

	logger.LogService(func() {
		zlog.Logger.Info().
			Msg("Getting all history")
	})

	if res, err = servs.repo.GetAllHistory(ctx, filters); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Msg("failed to get item history from database")
		})
		return nil, err
	}
	return res, nil
}

func validateItemRequest(req *models.ItemRequest) error {
	if req.Name == "" {
		return models.ErrItemEmptyName
	}
	if req.Price <= 0 {
		return models.ErrItemInvalidPrice
	}
	return nil
}

var _ interfaces.Service = (*WarehouseService)(nil)
