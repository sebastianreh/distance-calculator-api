package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sebastianreh/distance-calculator-api/internal/config"
	"github.com/sebastianreh/distance-calculator-api/pkg/logger"

	rd "github.com/go-redis/redis/v8"
)

const (
	emptyString = ""
)

type Redis interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type redis struct {
	client *rd.Client
}

func NewRedis(log logger.Logger, cfg config.Config) (Redis, error) {
	client := buildClient(cfg.Redis.Host)
	if client == nil {
		return nil, errors.New("error, connecting to redis server")
	}

	status := client.Ping(context.TODO())
	if status.Err() != nil {
		log.Error(fmt.Sprintf("server => %s Redis => Monitoring | Cannot connect to redis server | Error => %s",
			cfg.Redis.Host, status.Err()))
	} else {
		log.Info(fmt.Sprintf("Redis => Monitoring | Connected successfully to %s", cfg.Redis.Host), "")
	}

	return &redis{
		client: client,
	}, nil
}

func buildClient(address string) *rd.Client {
	var options = &rd.Options{
		PoolSize: 1000,
		OnConnect: func(ctx context.Context, cn *rd.Conn) error {
			return ctx.Err()
		},
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute * 5,
		Addr:         address,
	}

	client := rd.NewClient(options)
	return client
}

func (r *redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	status := r.client.Set(ctx, key, value, ttl)
	if status.Err() != nil {
		return status.Err()
	}

	return nil
}

func (r *redis) Get(ctx context.Context, key string) (string, error) {
	status := r.client.Get(ctx, key)
	if status.Err() != nil && status.Err() != rd.Nil {
		return emptyString, status.Err()
	}

	return status.Val(), nil
}
