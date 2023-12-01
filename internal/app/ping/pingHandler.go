package ping

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
)

type PingHandler interface {
	Ping(c echo.Context) error
}

type Response struct {
	Version string    `json:"version"`
	Name    string    `json:"name"`
	Uptime  time.Time `json:"uptime"`
}

type pingHandler struct {
	config config.Config
}

func NewPingHandler(cfg config.Config) PingHandler {
	return &pingHandler{
		config: cfg,
	}
}

func (s *pingHandler) Ping(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, Response{
		Version: s.config.ProjectVersion,
		Name:    s.config.ProjectName,
		Uptime:  time.Now().UTC(),
	})
}
