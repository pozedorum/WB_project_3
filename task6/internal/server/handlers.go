package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func (serv *SaleTrackerServer) CreateItem(c *ginext.Context) {
	var (
		req      models.SaleRequest
		saleInfo *models.SaleInformation
		err      error
	)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogServer(func() { zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for create event") })
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request format: " + err.Error()})
		return
	}
	if saleInfo, err = serv.service.CreateSale(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{
			"error":   "Failed to create sale",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, ginext.H{
		"id":          saleInfo.ID,
		"description": saleInfo.Description,
		"message":     "Sale created successfully",
	})
}

func (serv *SaleTrackerServer) GetItems(c *ginext.Context) {
	filters := make(map[string]interface{})

	// Парсим параметры запроса
	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filters["from"] = from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filters["to"] = to
		}
	}

	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}

	if saleType := c.Query("type"); saleType != "" {
		if saleType == "income" || saleType == "expense" {
			filters["type"] = saleType
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters["limit"] = limit
		}
	}

	sales, err := serv.service.GetAllSales(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{
			"error":   "Failed to get sales",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func (serv *SaleTrackerServer) GetItemByID(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid item ID"})
		return
	}

	sale, err := serv.service.GetSaleByID(c.Request.Context(), int64(id))
	if err != nil {
		if err == models.ErrSaleNotFound {
			c.JSON(http.StatusNotFound, ginext.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{
			"error":   "Failed to get sale",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sale)
}

func (serv *SaleTrackerServer) UpdateItem(c *ginext.Context) {
	var (
		req      models.SaleRequest
		saleInfo *models.SaleInformation
		err      error
	)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid ID parameter"})
		return
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if saleInfo, err = serv.service.UpdateSale(c.Request.Context(), int64(id), &req); err != nil {
		if err == models.ErrSaleNotFound {
			c.JSON(http.StatusNotFound, ginext.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{
			"error":   "Failed to update sale",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ginext.H{
		"id":          saleInfo.ID,
		"description": saleInfo.Description,
		"message":     "Sale updated successfully",
	})
}

func (serv *SaleTrackerServer) DeleteItem(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(models.StatusBadRequest, ginext.H{"error": models.ErrInvalidItemID.Error()})
		return
	}

	err = serv.service.DeleteSale(c.Request.Context(), int64(id))
	if err != nil {
		if err == models.ErrNegativeIndex {
			c.JSON(http.StatusBadRequest, ginext.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{
			"error":   "Failed to get sale",
			"details": err.Error(),
		})
		return
	}
}

func (serv *SaleTrackerServer) GetAnalytics(c *ginext.Context) {
	// Парсим query parameters
	req := parseAnalyticsRequest(c)
	if req == nil { // ошибка уже отправлена пользователю
		return
	}
	analytics, err := serv.service.GetAnalytics(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{
			"error":   "Failed to get analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (serv *SaleTrackerServer) ExportCSV(c *ginext.Context) {
	// Парсим query parameters
	req := parseAnalyticsRequest(c)
	if req == nil { // ошибка уже отправлена пользователю
		return
	}

	csvData, err := serv.service.ExportToCSV(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{
			"error":   "Failed to export CSV",
			"details": err.Error(),
		})
		return
	}

	// Устанавливаем заголовки для скачивания файла
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=sales_export_%s_to_%s.csv",
		req.From.Format("2006-01-02"), req.To.Format("2006-01-02")))
	c.Header("Content-Length", strconv.Itoa(len(csvData)))

	c.Data(http.StatusOK, "text/csv", csvData)
}

func parseAnalyticsRequest(c *ginext.Context) *models.AnalyticsRequest {
	var req models.AnalyticsRequest
	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			req.From = from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			req.To = to
		}
	}

	if req.From.IsZero() || req.To.IsZero() {
		c.JSON(http.StatusBadRequest, ginext.H{
			"error": models.ErrEmptyFromToDate.Error(),
		})
		return nil
	}

	if req.From.After(req.To) {
		c.JSON(http.StatusBadRequest, ginext.H{
			"error": models.ErrWrongTimeRange.Error(),
		})
		return nil
	}

	if category := c.Query("category"); category != "" {
		req.Category = category
	}

	if saleType := c.Query("type"); saleType != "" {
		if saleType == "income" || saleType == "expense" {
			req.Type = saleType
		} else {
			c.JSON(http.StatusBadRequest, ginext.H{
				"error": models.ErrInvalidType.Error(),
			})
			return nil
		}
	}

	if groupBy := c.Query("group_by"); groupBy != "" {
		if groupBy == "day" || groupBy == "week" || groupBy == "month" || groupBy == "category" {
			req.GroupBy = groupBy
		} else {
			c.JSON(http.StatusBadRequest, ginext.H{
				"error": models.ErrInvalidGroupBy.Error(),
			})
			return nil
		}
	}
	return &req
}
