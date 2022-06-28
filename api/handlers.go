package api

import (
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
	return c.JSON(http.StatusOK, task)
}
