package server

import (
	//"github.com/pozedorum/wbf/ginext"
	"fmt"

	"github.com/pozedorum/WB_project_3/task5/internal/models"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func (serv *EventBookerServer) CreateEvent(c *ginext.Context) {
	var req models.EventInformation
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for create event")
		c.JSON(models.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

}

func (serv *EventBookerServer) BookEvent(c *ginext.Context) {

}

func (serv *EventBookerServer) ConfirmBooking(c *ginext.Context) {

}

func (serv *EventBookerServer) GetEventInformation(c *ginext.Context) {

}

func parseProcessingOptions(c *ginext.Context) (*models.EventRequest, error) {
	infoStruct := &models.EventRequest{}
	// Вспомогательная функция для получения значения из form-data или query string
	getParam := func(key string) string {
		if val := c.PostForm(key); val != "" {
			return val
		}
		return c.Query(key)
	}
	// Парсим имя
	if infoStruct.Name = getParam("name"); infoStruct.Name == "" {
		return nil, fmt.Errorf("name of event cant be empty")
	}
	//if infoStruct.Date = time.Parse(layout string, value string)
	return infoStruct, nil
}
