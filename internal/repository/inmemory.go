package repository

import (
	"context"
	"sync"

	"github.com/voidroute/shortener/internal/domain"
)

type InMemoryRepository struct {
	mu    sync.RWMutex
	links map[string]domain.Link
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		links: make(map[string]domain.Link),
	}
}

func (r *InMemoryRepository) Save(_ context.Context, link *domain.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[link.Code]; exists {
		return ErrCodeExists
	}

	r.links[link.Code] = *link
	return nil
}

func (r *InMemoryRepository) Get(_ context.Context, code string) (*domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, exists := r.links[code]
	if !exists {
		return nil, ErrLinkNotFound
	}

	return &link, nil
}

func (r *InMemoryRepository) Delete(_ context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.links[code]
	if !exists {
		return ErrLinkNotFound
	}

	delete(r.links, code)
	return nil
}

func (r *InMemoryRepository) Ping(_ context.Context) error {
	return nil
}

func (r *InMemoryRepository) Close() error {
	return nil
}
