package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type redisClient struct {
	redis *redis.Client
}

func NewRedisClient(addr string, password string) *redisClient {
	return &redisClient{
		redis: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0, // use default DB
		}),
	}
}

func (c *redisClient) Set(ctx context.Context, k string, v any, d time.Duration) error {
	if err := c.redis.Set(ctx, k, v, d).Err(); err != nil {
		return fmt.Errorf("%w: set: %v", ErrInternal, err)
	}
	return nil
}

func (c *redisClient) Get(ctx context.Context, k string) (any, error) {
	v, err := c.redis.Get(ctx, k).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: get: %v", ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w: get: %v", ErrInternal, err)
	}

	return v, nil
}
