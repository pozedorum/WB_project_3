package server

import (
	"github.com/pozedorum/WB_project_3/task5/internal/models"
	"github.com/pozedorum/WB_project_3/task5/internal/service"
	"github.com/pozedorum/WB_project_3/task5/pkg/config"
	"github.com/wb-go/wbf/ginext"
)

type EventBookerServer struct {
	service   *service.EventBookerService
	jwtConfig *config.JWTConfig
}

func New(service *service.EventBookerService, jwtConfig *config.JWTConfig) *EventBookerServer {
	return &EventBookerServer{service: service, jwtConfig: jwtConfig}
}

func (serv *EventBookerServer) SetupRoutes(router *ginext.RouterGroup) {
	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())
	router.Use(CORSMiddleware())

	// Обслуживаем статические файлы фронтенда
	router.Static("/static", "./internal/frontend/static")

	// Главная страница
	router.GET("/", serv.ServeFrontend)

	// API routes
	router.POST("/register", serv.RegisterUser)
	router.POST("/login", serv.LoginUser)
	router.GET("/events", serv.GetAllEvents)
	router.GET("/events/:id", serv.GetEventInformation)

	// Protected routes
	protected := router.Group("/")
	protected.Use(serv.JWTAuthMiddleware())
	{
		protected.POST("/events", serv.CreateEvent)
		protected.POST("/events/:id/book", serv.BookEvent)
		protected.POST("/events/:id/confirm", serv.ConfirmBooking)
	}
}

func (serv *EventBookerServer) ServeFrontend(c *ginext.Context) {
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
