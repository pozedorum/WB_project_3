package server

import (
	"net/http"

	"github.com/pozedorum/WB_project_3/task6/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/wb-go/wbf/ginext"
)

type SaleTrackerServer struct {
	service interfaces.SaleService
}

func New(service interfaces.SaleService) *SaleTrackerServer {
	return &SaleTrackerServer{service: service}
}

func (serv *SaleTrackerServer) SetupRoutes(router *ginext.Engine, apiRouter *ginext.RouterGroup) {
	apiRouter.Use(ginext.Logger())
	apiRouter.Use(ginext.Recovery())
	apiRouter.Use(CORSMiddleware())

	// Обслуживаем статические файлы фронтенда и загружаем страницы
	apiRouter.Static("/static", "internal/frontend/static")
	router.LoadHTMLGlob("internal/frontend/templates/*.html")

	// Страницы
	apiRouter.GET("/", serv.ServeFrontend)
	apiRouter.GET("/analytics", serv.ServeAnalytics) // Страница аналитики

	// API routes с префиксом /api
	api := apiRouter.Group("/api")
	{
		api.POST("/items", serv.CreateItem)
		api.GET("/items", serv.GetItems)
		api.GET("/items/:id", serv.GetItemByID)
		api.PUT("/items/:id", serv.UpdateItem)
		api.DELETE("/items/:id", serv.DeleteItem)
		api.GET("/analytics", serv.GetAnalytics) // API: /api/analytics
		api.GET("/csv", serv.ExportCSV)          // API: /api/csv
	}
}

func (serv *SaleTrackerServer) ServeFrontend(c *ginext.Context) {
	c.HTML(models.StatusOK, "index.html", nil)
}

func (s *SaleTrackerServer) ServeAnalytics(c *ginext.Context) {
	c.HTML(http.StatusOK, "analytics.html", ginext.H{
		"title": "Sales Analytics",
	})
}

func CORSMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
