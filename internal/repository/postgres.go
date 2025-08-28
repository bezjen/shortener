package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (p *PostgresRepository) Save(ctx context.Context, userID string, url model.URL) error {
	_, err := p.db.ExecContext(ctx,
		"insert into t_short_url(short_url, original_url, user_id) values ($1, $2, $3)"+
			"on conflict (original_url) do update "+
			"set short_url = EXCLUDED.short_url, user_id = EXCLUDED.user_id, is_deleted = false "+
			"where t_short_url.is_deleted = true",
		url.ShortURL, url.OriginalURL, userID)
	if err != nil {
		if isUniqueViolation(err) {
			uniqueErr, errConflict := p.newURLExistsError(ctx, url.OriginalURL)
			if errConflict != nil {
				return errConflict
			}
			return uniqueErr
		}
		return err
	}
	return nil
}

func (p *PostgresRepository) SaveBatch(ctx context.Context, userID string, urls []model.URL) error {
	tx, err := p.db.Begin()

	if err != nil {
		return err
	}

	for _, url := range urls {
		_, err = tx.ExecContext(ctx,
			"insert into t_short_url(short_url, original_url, user_id) values ($1, $2, $3)"+
				"on conflict (original_url) do update "+
				"set short_url = EXCLUDED.short_url, user_id = EXCLUDED.user_id, is_deleted = false "+
				"where t_short_url.is_deleted = true",
			url.ShortURL, url.OriginalURL, userID)
		if err != nil {
			errRollback := tx.Rollback()
			if errRollback != nil {
				return errRollback
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

func (p *PostgresRepository) GetByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	rows, err := p.db.QueryContext(ctx,
		"select short_url, original_url from t_short_url where user_id = $1",
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs for user %s: %w", userID, err)
	}
	defer rows.Close()
	var urls []model.URL
	for rows.Next() {
		var url model.URL
		err = rows.Scan(&url.ShortURL, &url.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return urls, nil
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
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
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
