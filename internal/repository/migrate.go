package repository

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Migrate(repo *PostgresRepository, migrationsDir string) error {
	db := stdlib.OpenDBFromPool(repo.pool)
	defer func(db *sql.DB) {
		err := db.Close()

		if err != nil {
			slog.Error("failed to close migration db connection", "error", err)
		}
	}(db)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
