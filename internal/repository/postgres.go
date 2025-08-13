package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/bezjen/shortener/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(databaseDSN string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		db: db,
	}, nil
}

func (p *PostgresRepository) Save(ctx context.Context, url model.URL) error {
	_, err := p.db.ExecContext(ctx,
		"insert into t_short_url(short_url, original_url) values ($1, $2)", url.ShortURL, url.OriginalURL)
	if err != nil {
		if isUniqueViolation(err) {
			uniqueErr, err := p.newURLExistsError(ctx, url.OriginalURL)
			if err != nil {
				return err
			}
			return uniqueErr
		}
		return err
	}
	return nil
}

func (p *PostgresRepository) SaveBatch(ctx context.Context, urls []model.URL) error {
	tx, err := p.db.Begin()

	if err != nil {
		return err
	}

	for _, url := range urls {
		_, err = tx.ExecContext(ctx,
			"insert into t_short_url(short_url, original_url) values ($1, $2)", url.ShortURL, url.OriginalURL)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
	}

	return tx.Commit()
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

func (p *PostgresRepository) newURLExistsError(ctx context.Context, originalURL string) (*ErrURLConflict, error) {
	var shortURL string
	row := p.db.QueryRowContext(ctx, "select short_url from t_short_url where original_url = $1;", originalURL)
	err := row.Scan(&shortURL)
	if err != nil {
		return nil, err
	}
	return &ErrURLConflict{ShortURL: shortURL, Err: "Original URL already exists"}, nil
}

func (p *PostgresRepository) getShortURLByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	row := p.db.QueryRowContext(ctx, "select short_url from t_short_url where original_url = $1", originalURL)
	var shortUrl string
	err := row.Scan(&shortUrl)
	if err != nil {
		return "", err
	}
	return shortUrl, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation
	}
	return false
}

type ErrURLConflict struct {
	ShortURL string
	Err      string
}

func (err *ErrURLConflict) Error() string {
	return err.Err
}
