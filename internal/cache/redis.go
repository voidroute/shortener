package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/voidroute/shortener/internal/domain"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr string, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *RedisCache) Set(ctx context.Context, code string, link *domain.Link) error {
	data, err := json.Marshal(link)
	if err != nil {
		return fmt.Errorf("failed to marshal link: %w", err)
	}

	if err = c.client.Set(ctx, code, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (c *RedisCache) Get(ctx context.Context, code string) (*domain.Link, error) {
	data, err := c.client.Get(ctx, code).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	var link domain.Link
	if err = json.Unmarshal(data, &link); err != nil {
		return nil, fmt.Errorf("failed to unmarshal link: %w", err)
	}

	return &link, nil
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %w", err)
	}
	return nil
}
