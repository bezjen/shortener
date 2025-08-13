package repository

import (
	"database/sql"
	"github.com/bezjen/shortener/internal/logger"
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
