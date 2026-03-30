package cache

import (
	"context"

	"github.com/voidroute/shortener/internal/domain"
)

type Cache interface {
	Get(ctx context.Context, code string) (*domain.Link, error)
	Set(ctx context.Context, code string, link *domain.Link) error
	Ping(ctx context.Context) error
	Close() error
}
