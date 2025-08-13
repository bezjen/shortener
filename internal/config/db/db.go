package db

import (
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/golang-migrate/migrate/v4"
)

func InitDB(logger *logger.Logger, cfg config.Config) (repository.Repository, error) {
	if cfg.DatabaseDSN != "" {
		if err := runMigrations(cfg.DatabaseDSN); err != nil {
			return nil, err
		}

		repoDB, err := repository.NewPostgresRepository(logger, cfg.DatabaseDSN)
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
