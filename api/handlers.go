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
	task.TaskID = s.getNewTaskId()
	if err := c.Validate(task); err != nil {
		return err
	}

	routingKey := task.QueueID + "_rKey"
	data, err := json.Marshal(task)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = s.mqProducer.PublishMessage(routingKey, data, task.Priority)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	err = s.pos.AddItem(task.TaskID, int(s.mqProducer.MaxPriority-task.Priority))
	p, err := s.pos.GetPosition(task.TaskID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	task.Position = p + 1
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

func (s *Server) getNewTaskId() string {
	id := uuid.New().String()
	id = fmt.Sprintf("%d_%s", s.keyCounter, id)
	s.keyCounter++
	return id
}
