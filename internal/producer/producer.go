package producer

import (
	"context"

	"github.com/voidroute/shortener/internal/domain"
)

type ClickProducer interface {
	SendClickEvent(ctx context.Context, link *domain.Link, ip string) error
	Close() error
}
