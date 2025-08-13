package repository

import (
	"database/sql"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/model"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresRepository struct {
	logger *logger.Logger
	db     *sql.DB
}

func NewPostgresRepository(logger *logger.Logger, databaseDSN string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		logger: logger,
		db:     db,
	}, nil
}

func (p *PostgresRepository) Save(shortURL string, url string) error {
	_, err := p.buildShortURLDto(shortURL, url)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRepository) GetByShortURL(shortURL string) (string, error) {
	return "", nil
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}

func (p *PostgresRepository) Ping() error {
	return p.db.Ping()
}

func (p *PostgresRepository) buildShortURLDto(shortURL string, originalURL string) (*model.ShortURLDto, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	shortURLDto := model.ShortURLDto{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	return &shortURLDto, nil
}
