package config

import (
	"github.com/kelseyhightower/envconfig"
)

type (
	Config struct {
		ProjectName    string `default:"distance-calculator-api"`
		ProjectVersion string `envconfig:"PROJECT_VERSION" default:"0.0.1"`
		Port           string `envconfig:"PORT" default:"8000" required:"true"`
		Prefix         string `envconfig:"PREFIX" default:"/distance-calculator-api"`
		Env            string `envconfig:"ENV" default:"prod"`
		Redis          struct {
			Host string `envconfig:"REDIS_HOST" default:"127.0.0.1:6379"`
		}
		MaxDeliveryRadius float64 `envconfig:"MAX_DELIVERY_RADIUS" default:"6"`
	}
)

var (
	Configs Config
)

func NewConfig() Config {
	if err := envconfig.Process("", &Configs); err != nil {
		panic(err.Error())
	}

	return Configs
}
