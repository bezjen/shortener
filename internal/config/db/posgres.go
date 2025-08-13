package db

import (
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/repository"
)

func InitDB(logger *logger.Logger, cfg config.Config) (repository.Repository, error) {
	if cfg.DatabaseDSN != "" {
		repoDB, err := repository.NewPostgresRepository(logger, cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return repoDB, nil
	} else {
		repoFile, err := repository.NewFileRepository(cfg)
		if err != nil {
			return nil, err
		}
		return repoFile, nil
	}
}
