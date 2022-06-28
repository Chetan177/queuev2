package api

import (
	"github.com/labstack/echo/v4"
	"log"
)

type url struct {
	Path    string
	Handler func(echo.Context) error
	Method  string
}

func (s *Server) loadQueueGroup() {
	queueGroup := s.restServer.Group(s.getAccountLevelBaseURL() + "/task")
	routes := s.getTaskRoutes()
	s.loadRoutes(queueGroup, routes)
}

func (s *Server) getAccountLevelBaseURL() string {
	return "/v1.0/accounts/:accountID"
}

func (s *Server) loadRoutes(group *echo.Group, routes []url) {

	log.Println("Loading routes ", routes)

	for _, route := range routes {
		switch route.Method {
		case "DELETE":
			group.DELETE(route.Path, route.Handler)
		case "GET":
			group.GET(route.Path, route.Handler)
		case "POST":
			group.POST(route.Path, route.Handler)
		case "PUT":
			group.PUT(route.Path, route.Handler)
		}
	}
}

func (s *Server) getTaskRoutes() []url {
	Urls := []url{

		{"", s.submitTask, "POST"},
	}

	return Urls
}

func (s *Server) getQueueRoutes() []url {
	Urls := []url{

		{"", s.submitTask, "POST"},
	}

	return Urls
}