package repository

import (
	"log/slog"

	"github.com/go-redis/redis"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg"
)

type Redis struct {
	client *redis.Client

	logger *slog.Logger
}

var _ pkg.Repository = (*Redis)(nil)

func newRedis(
	address, password string, db int,

	logger *slog.Logger,
) pkg.Repository {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
			DB:       db,
		}),
		logger: logger,
	}
}

// TODO: add logging here
func (r *Redis) SetMessageID(key string, value string) error {
	return r.client.Set(key, value, 0).Err()
}

func (r *Redis) GetMessageID(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *Redis) DeleteMessageID(key string) error {
	return r.client.Del(key).Err()
}

func (r *Redis) Close() error {
	return r.client.Close()
}
