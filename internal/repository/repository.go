package repository

import (
	"context"

	"github.com/voidroute/shortener/internal/domain"
)

type LinkRepository interface {
	Save(ctx context.Context, link *domain.Link) error
	Get(ctx context.Context, code string) (*domain.Link, error)
	Delete(ctx context.Context, code string) error
	Ping(ctx context.Context) error
	Close() error
}
