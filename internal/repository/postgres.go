package repository

import (
	"context"
	"github.com/bezjen/shortener/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(databaseDSN string) (*PostgresRepository, error) {
	db, err := sqlx.Connect("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (p *PostgresRepository) Save(ctx context.Context, url model.URL) error {
	_, err := p.db.NamedExecContext(ctx,
		"insert into t_short_url(short_url, original_url) values (:short_url, :original_url)", url)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRepository) SaveBatch(ctx context.Context, urls []model.URL) error {
	tx, err := p.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	for _, url := range urls {
		_, err = tx.NamedExecContext(ctx,
			"insert into t_short_url(short_url, original_url) values (:short_url, :original_url)", url)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return rbErr
			}
			return err
		}
	}

	return tx.Commit()
}

func (p *PostgresRepository) GetByShortURL(ctx context.Context, shortURL string) (originalUrl string, err error) {
	err = p.db.GetContext(ctx, &originalUrl,
		"select original_url from t_short_url where short_url = $1", shortURL)
	return
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}
