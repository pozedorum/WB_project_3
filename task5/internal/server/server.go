package server

import (
	"github.com/pozedorum/WB_project_3/task5/internal/service"
	"github.com/wb-go/wbf/ginext"
)

type EventBookerServer struct {
	service service.EventBookerService
}

func New(service service.EventBookerService) *EventBookerServer {
	return &EventBookerServer{service: service}
}

func (serv *EventBookerServer) SetupRoutes(router *ginext.RouterGroup) {
	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())
	router.Use(CORSMiddleware())

	// // Статические файлы
	// router.Static("/static", "./internal/frontend/static")

	// // Главная страница
	// router.GET("/", serv.ServeFrontend)

	router.POST("/events", serv.CreateEvent)
	router.POST("/events/:id/book", serv.BookEvent)
	router.POST("/events/:id/confirm", serv.ConfirmBooking)
	router.GET("/events/:id", serv.GetEventInformation)
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
