package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"queuev2/mq/producer"
	"regexp"
	"sync"
)

var (
	once   sync.Once
	server *Server
)

type Server struct {
	restServerPort int
	restServer     *echo.Echo
	mqProducer     *producer.MQProducer
}

type CustomValidator struct {
	validator *validator.Validate
}

func GetServerInstance(restPort int, mqProducer *producer.MQProducer) *Server {
	once.Do(func() {
		server = createServer(restPort, mqProducer)
	})

	return server
}

func createServer(restPort int, mqProducer *producer.MQProducer) *Server {

	apiServer := echo.New()
	apiServer.Use(middleware.Recover())
	apiServer.Pre(middleware.RemoveTrailingSlash())
	apiServer.Use(middleware.BodyDump(func(e echo.Context, reqBody []byte, respBody []byte) {
		bodyDumpHandler(e, reqBody, respBody)
	}))
	apiServer.Validator = &CustomValidator{validator: validator.New()}
	apiServer.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	apiServer.GET("/health", healthCheck)
	//apiServer.GET("/swagger/*", echoSwagger.WrapHandler)
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(apiServer)

	s := &Server{
		restServerPort: restPort,
		restServer:     apiServer,
		mqProducer:     mqProducer,
	}

	return s
}

func (s *Server) StartServer() error {

	s.loadQueueGroup()
	s.loadTaskGroup()

	data, err := json.MarshalIndent(s.restServer.Routes(), "", "  ")
	if err != nil {
		return err
	}
	log.Println("Echo routes loaded: ", string(data))

	//start rest server
	go func(port int) {
		addr := fmt.Sprintf(":%d", port)
		s.restServer.Logger.Fatal(s.restServer.Start(addr))
	}(s.restServerPort)

	return nil
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "QueueService API is up and running")
}

func bodyDumpHandler(e echo.Context, reqBody, resBody []byte) {
	if e.Request().URL.Path == "/metrics" {
		return
	}
	accID := e.Param("accountID")
	uuidWithHash := e.Param("request_uuid")
	reqBodyString, resBodyString := preprocessorBeforeBodyDump(string(reqBody), string(resBody))
	log.Println(uuidWithHash, e.Request().URL.Path, fmt.Sprintf("AccountID: %v, Request Body: %+v, Response: %+v", accID, reqBodyString, resBodyString))

}
func preprocessorBeforeBodyDump(reqBody, resBody string) (string, string) {
	//removing newline \r\n, backslash and double whitespace added in the process of req/res bodydumping
	re := regexp.MustCompile(newLineAndForwardSlashRegEx)
	reDWSpace := regexp.MustCompile(doubleWhiteSpaceRegEx)
	reqBody = re.ReplaceAllString(reqBody, "")
	reqBody = reDWSpace.ReplaceAllString(reqBody, " ")
	resBody = re.ReplaceAllString(resBody, "")
	resBody = reDWSpace.ReplaceAllString(resBody, " ")
	return reqBody, resBody
}
