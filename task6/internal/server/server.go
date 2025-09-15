package server

import (
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
	apiRouter.Static("/static", "./internal/frontend/static")
	router.LoadHTMLGlob("internal/frontend/templates/*.html")
	// Главная страница
	apiRouter.GET("/", serv.ServeFrontend)

	// API routes
	apiRouter.POST("/items", serv.CreateItem)
	apiRouter.GET("/items", serv.GetItems)
	apiRouter.GET("/items/:id", serv.GetItemByID)
	apiRouter.PUT("/items/:id", serv.UpdateItem)
	apiRouter.DELETE("/items/:id", serv.DeleteItem)

	apiRouter.GET("/analytics", serv.GetAnalytics)
	apiRouter.GET("/csv", serv.ExportCSV)

}

func (serv *SaleTrackerServer) ServeFrontend(c *ginext.Context) {
	c.HTML(models.StatusOK, "index.html", nil)
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
