package main

import (
	"github.com/sebastianreh/distance-calculator-api/cmd/httpserver"
	"github.com/sebastianreh/distance-calculator-api/internal/container"
)

func main() {
	dependencies := container.Build()
	server := httpserver.NewServer(dependencies)
	server.Middlewares(httpserver.WithRecover(),
		httpserver.WithLogger(dependencies.Config),
	)
	server.Routes()
	server.SetErrorHandler(httpserver.HTTPErrorHandler)
	server.Start()
}
