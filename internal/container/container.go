package container

import (
	"github.com/go-resty/resty/v2"
	"github.com/sebastianreh/distance-calculator-api/internal/app/calculator"
	"github.com/sebastianreh/distance-calculator-api/internal/app/ping"
	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"
	rds "github.com/sebastianreh/distance-calculator-api/pkg/redis"
	"github.com/sebastianreh/distance-calculator-api/pkg/rest"
)

type Dependencies struct {
	PingHandler       ping.PingHandler
	Config            config.Config
	Logs              logger.Logger
	CalculatorHandler calculator.CalculatorHandler
}

func Build() Dependencies {
	dependencies := Dependencies{}
	dependencies.Config = config.NewConfig()
	logs := logger.NewLogger()
	dependencies.Logs = logs
	dependencies.PingHandler = ping.NewPingHandler(dependencies.Config)

	redis, err := rds.NewRedis(logs, dependencies.Config)
	if err != nil {
		logs.Fatal(err.Error())
	}

	restyClient := resty.New()
	s3RestClient := rest.NewS3Client(logs, restyClient)
	calculatorRepository := calculator.NewCalculatorRepository(dependencies.Config, redis, logs)
	calculatorService := calculator.NewCalculatorService(dependencies.Config, calculatorRepository, s3RestClient, logs)
	calculatorHandler := calculator.NewCalculatorHandler(dependencies.Config, calculatorService, logs)

	dependencies.CalculatorHandler = calculatorHandler

	return dependencies
}
