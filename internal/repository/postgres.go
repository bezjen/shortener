package repository

import (
	"context"
	"database/sql"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/model"
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

func (p *PostgresRepository) Save(ctx context.Context, url model.URL) error {
	_, err := p.db.ExecContext(ctx,
		"insert into t_short_url(short_url, original_url) values ($1, $2)", url.ShortURL, url.OriginalURL)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRepository) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
	row := p.db.QueryRowContext(ctx, "select original_url from t_short_url where short_url = $1", shortURL)
	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}
