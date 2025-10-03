package di

import (
	"net/http"

	"github.com/pozedorum/WB_project_3/task7/internal/interfaces"
	"github.com/wb-go/wbf/dbpg"
)

type Container struct {
	db         *dbpg.DB
	httpServer *http.Server

	repo    interfaces.ItemRepository
	service interfaces.ItemService
	server  interfaces.WarehouseServer

	closers []interfaces.Closer // Список ресурсов, которые нужно закрыть
}
