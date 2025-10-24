// Package db provides database initialization and migration functionality.
// It handles the setup of different storage backends (PostgreSQL, file, in-memory)
// and runs database migrations for PostgreSQL.
package db

import (
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// InitDB initializes the appropriate repository based on configuration.
// It supports PostgreSQL, file-based, and in-memory storage backends.
// For PostgreSQL, it runs database migrations before initializing the repository.
//
// Parameters:
//   - cfg: application configuration containing storage settings
//
// Returns:
//   - repository.Repository: initialized repository instance
//   - error: error if initialization or migrations fail
//
// Example:
//
//	cfg := config.Config{DatabaseDSN: "postgres://user:pass@localhost/db"}
//	repo, err := db.InitDB(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
func InitDB(cfg config.Config) (repository.Repository, error) {
	if cfg.DatabaseDSN != "" {
		if err := runMigrations(cfg.DatabaseDSN); err != nil {
			return nil, err
		}

		repoDB, err := repository.NewPostgresRepository(cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return repoDB, nil
	} else if cfg.FileStoragePath != "" {
		repoFile, err := repository.NewFileRepository(cfg)
		if err != nil {
			return nil, err
		}
		return repoFile, nil
	} else {
		return repository.NewInMemoryRepository(), nil
	}
}

// runMigrations executes database migrations for PostgreSQL.
// It looks for migration files in the "migrations" directory and applies them.
// If no new migrations are found (ErrNoChange), it returns nil.
//
// Parameters:
//   - databaseDSN: PostgreSQL connection string
//
// Returns:
//   - error: error if migrations fail (excluding no-change errors)
func runMigrations(databaseDSN string) error {
	m, err := migrate.New("file://migrations", databaseDSN)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
