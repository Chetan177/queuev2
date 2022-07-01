package api

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (s *Server) submitTask(c echo.Context) error {
	task := new(Task)
	if err := c.Bind(task); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	task.TaskID = uuid.New().String()
	if err := c.Validate(task); err != nil {
		return err
	}

	routingKey := task.QueueID + "_rKey"
	data, err := json.Marshal(task)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	s.mqProducer.PublishMessage(routingKey, data, task.Priority)

	return c.JSON(http.StatusOK, task)
}

func (s *Server) createQueue(c echo.Context) error {
	queue := new(Queue)
	accID := c.Param("accountID")
	if err := c.Bind(queue); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	queue.QueueID = generateQueueID(accID, queue.QueueName, queue.MaxPriority)

	if err := c.Validate(queue); err != nil {
		return err
	}

	err := s.mqProducer.CreateQueue(queue.QueueID, queue.MaxPriority)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, queue)
}

func generateQueueID(accID, queueName string, priority uint8) string {
	qID := accID + "_" + queueName + fmt.Sprintf("%d", priority)
	qEnc := b64.StdEncoding.EncodeToString([]byte(qID))
	return "QID_" + qEnc
}
