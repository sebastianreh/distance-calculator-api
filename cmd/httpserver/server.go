package httpserver

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sebastianreh/distance-calculator-api/internal/container"
)

type Server struct {
	Server       *echo.Echo
	dependencies container.Dependencies
}

func NewServer(dependencies container.Dependencies) *Server {
	return &Server{
		Server:       echo.New(),
		dependencies: dependencies,
	}
}

// Start run the server
func (s *Server) Start() {
	s.Server.Logger.Fatal(s.Server.Start(fmt.Sprintf(":%s", s.dependencies.Config.Port)))
}

func (s *Server) SetErrorHandler(errorHandler echo.HTTPErrorHandler) {
	s.Server.HTTPErrorHandler = errorHandler
}

func (s *Server) NewServerContext(request *http.Request, writer http.ResponseWriter) echo.Context {
	return s.Server.NewContext(request, writer)
}
