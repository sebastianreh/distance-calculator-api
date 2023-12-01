package httpserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sebastianreh/distance-calculator-api/cmd/httpserver/resterror"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/internal/entities/exceptions"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Middleware func(*Server)

// Middlewares build the middlewares of the server
func (s *Server) Middlewares(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		middleware(s)
	}
}

func WithLogger(cfg config.Config) Middleware {
	return func(s *Server) {
		s.Server.Use(echoMiddleware.LoggerWithConfig(echoMiddleware.LoggerConfig{
			Skipper: func(e echo.Context) bool {
				return strings.Contains(e.Path(), "ping")
			},
			CustomTimeFormat: "2006-01-02T15:04:05.1483386-00:00",
			Format: `{ "time":"${time_custom}", "level" :"Info" ,"method":"${method}", "uri":"${uri}",` +
				fmt.Sprintf(`"service": "%q" }`,
					cfg.ProjectName) + "\n",
		}))
	}
}

func WithRecover() Middleware {
	return func(s *Server) {
		s.Server.Use(echoMiddleware.Recover())
	}
}

func HTTPErrorHandler(err error, ctx echo.Context) {
	var apiError resterror.RestErr
	switch value := err.(type) {
	case *echo.HTTPError:
		apiError = resterror.NewRestError(value.Error(), value.Code, err.Error())
	case exceptions.DuplicatedException:
		apiError = resterror.NewRestError(err.Error(), http.StatusConflict, "conflict")
	case resterror.RestErr:
		apiError = value
	case exceptions.NotFoundException:
		apiError = resterror.NewNotFoundError(err.Error())
	case exceptions.UnauthorizedException:
		apiError = resterror.NewUnauthorizedError(err.Error())
	default:
		apiError = resterror.NewInternalServerError(err.Error(), err)
	}

	ctx.JSON(apiError.Status(), apiError)
}
