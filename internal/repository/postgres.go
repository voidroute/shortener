package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voidroute/shortener/internal/config"
	"github.com/voidroute/shortener/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(cfg config.DatabaseConfig) (*PostgresRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime

	return newFromPoolConfig(poolCfg)
}

func NewPostgresRepositoryFromDSN(dsn string) (*PostgresRepository, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	return newFromPoolConfig(poolCfg)
}

func newFromPoolConfig(poolCfg *pgxpool.Config) (*PostgresRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) Save(ctx context.Context, link *domain.Link) error {
	query := `INSERT INTO links (code, url, created_at, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, link.Code, link.URL, link.CreatedAt, link.ExpiresAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrCodeExists
		}

		return fmt.Errorf("failed to save link: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, code string) (*domain.Link, error) {
	query := `SELECT code, url, created_at, expires_at FROM links WHERE code = $1`
	row := r.pool.QueryRow(ctx, query, code)

	var link domain.Link
	if err := row.Scan(&link.Code, &link.URL, &link.CreatedAt, &link.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	link.CreatedAt = link.CreatedAt.UTC()
	if link.ExpiresAt != nil {
		utc := link.ExpiresAt.UTC()
		link.ExpiresAt = &utc
	}

	return &link, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, code string) error {
	query := `DELETE FROM links WHERE code = $1`

	result, err := r.pool.Exec(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrLinkNotFound
	}

	return nil
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}
