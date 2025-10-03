package server

import (
	"net/http"

	"github.com/pozedorum/WB_project_3/task7/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task7/internal/models"
	"github.com/wb-go/wbf/ginext"
)

type WarehouseServer struct {
	jwtConfig JWTConfig
	service   interfaces.ItemService
}

func New(service interfaces.ItemService) *WarehouseServer {
	return &WarehouseServer{service: service}
}

func (serv *WarehouseServer) SetupRoutes(router *ginext.Engine, apiRouter *ginext.RouterGroup) {
	apiRouter.Use(ginext.Logger())
	apiRouter.Use(ginext.Recovery())
	apiRouter.Use(CORSMiddleware())

	// Обслуживаем статические файлы фронтенда и загружаем страницы
	apiRouter.Static("/static", "internal/frontend/static")
	router.LoadHTMLGlob("internal/frontend/templates/*.html")

	// Страницы
	apiRouter.GET("/", serv.ServeFrontend)

	// Публичные роуты
	apiRouter.POST("/login", serv.Login)
	apiRouter.GET("/profile", serv.JWTAuthMiddleware(), serv.GetProfile)

	// Защищенные роуты для товаров
	items := apiRouter.Group("/items")
	items.Use(serv.JWTAuthMiddleware())

	// Все роли могут просматривать
	items.GET("", serv.GetItems)
	items.GET("/:id", serv.GetItemByID)

	// Только admin и manager могут создавать/изменять/удалять
	items.POST("", serv.RoleMiddleware(models.RoleAdmin, models.RoleManager), serv.CreateItem)
	items.PUT("/:id", serv.RoleMiddleware(models.RoleAdmin, models.RoleManager), serv.UpdateItem)
	items.DELETE("/:id", serv.RoleMiddleware(models.RoleAdmin, models.RoleManager), serv.DeleteItem)

	// История изменений
	history := apiRouter.Group("/history")
	history.Use(serv.JWTAuthMiddleware())

	// Все роли могут смотреть историю конкретного товара
	history.GET("/item/:id", serv.GetItemHistory)

	// Только admin и auditor могут смотреть всю историю
	history.GET("", serv.RoleMiddleware(models.RoleAdmin, models.RoleViewer), serv.GetAllHistory)
	history.GET("/export", serv.RoleMiddleware(models.RoleAdmin, models.RoleViewer), serv.ExportHistory)
}

func (serv *WarehouseServer) ServeFrontend(c *ginext.Context) {
	c.HTML(models.StatusOK, "index.html", nil)
}

func (s *WarehouseServer) ServeAnalytics(c *ginext.Context) {
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

var _ interfaces.WarehouseServer = (*WarehouseServer)(nil)
