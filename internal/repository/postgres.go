// Package repository provides data storage implementations for the URL shortening service.
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

// PostgresRepository implements Repository interface for PostgreSQL storage.
// It provides persistent storage with transaction support and concurrent access.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository instance.
// Initializes database connection using the provided connection string.
//
// Parameters:
//   - databaseDSN: PostgreSQL connection string
//
// Returns:
//   - *PostgresRepository: initialized PostgreSQL repository
//   - error: error if database connection fails
func NewPostgresRepository(databaseDSN string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		db: db,
	}, nil
}

// Save stores a URL mapping in PostgreSQL database.
// Handles unique constraint violations and returns appropriate errors.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: identifier of the user creating the URL
//   - url: URL object containing short and original URLs
//
// Returns:
//   - error: error if database operation fails or URL conflict occurs
func (p *PostgresRepository) Save(ctx context.Context, userID string, url model.URL) error {
	_, err := p.db.ExecContext(ctx,
		"insert into t_short_url(short_url, original_url, user_id, is_deleted) values ($1, $2, $3, false)",
		url.ShortURL, url.OriginalURL, userID)
	if err != nil {
		if isUniqueViolation(err) {
			var shortURL string
			var isDeleted bool
			row := p.db.QueryRowContext(ctx,
				"select short_url, is_deleted from t_short_url where original_url = $1;",
				url.OriginalURL)
			errScan := row.Scan(&shortURL, &isDeleted)
			if errScan != nil {
				return errScan
			}
			if isDeleted {
				_, errUpdate := p.db.ExecContext(ctx,
					"update t_short_url set short_url = $1, user_id = $2, is_deleted = false where original_url = $3;",
					url.ShortURL, userID, url.OriginalURL)
				return errUpdate
			}
			return &ErrURLConflict{ShortURL: shortURL, Err: "Original URL already exists"}
		}
		return err
	}
	return nil
}

// SaveBatch stores multiple URL mappings in a single transaction.
// Provides atomic batch operations - all URLs are saved or none.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: identifier of the user creating the URLs
//   - urls: slice of URL objects to store
//
// Returns:
//   - error: error if any database operation fails
func (p *PostgresRepository) SaveBatch(ctx context.Context, userID string, urls []model.URL) error {
	if len(urls) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, url := range urls {
		_, err = p.db.ExecContext(ctx,
			"insert into t_short_url(short_url, original_url, user_id, is_deleted) values ($1, $2, $3, false)",
			url.ShortURL, url.OriginalURL, userID)
		if err != nil {
			if isUniqueViolation(err) {
				var isDeleted bool
				row := p.db.QueryRowContext(ctx,
					"select is_deleted from t_short_url where original_url = $1;",
					url.OriginalURL)
				errScan := row.Scan(&isDeleted)
				if errScan != nil {
					errRollback := tx.Rollback()
					if errRollback != nil {
						return errRollback
					}
					return errScan
				}
				if isDeleted {
					_, errUpdate := p.db.ExecContext(ctx,
						"update t_short_url set short_url = $1, user_id = $2, is_deleted = false where original_url = $3;",
						url.ShortURL, userID, url.OriginalURL)
					if errUpdate == nil {
						continue
					}
				}
			}
			errRollback := tx.Rollback()
			if errRollback != nil {
				return errRollback
			}
			return err
		}
	}

	return tx.Commit()
}

// DeleteBatch marks multiple short URLs as deleted in a single transaction.
// Uses PostgreSQL array parameter for efficient batch updates.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: identifier of the user owning the URLs (currently not used for authorization)
//   - shortURLs: slice of short URL identifiers to mark as deleted
//
// Returns:
//   - error: error if database operation fails
func (p *PostgresRepository) DeleteBatch(ctx context.Context, _ string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "update t_short_url set is_deleted = true where short_url = any($1::text[])")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, shortURLs)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetByShortURL retrieves the original URL by its short identifier.
// Returns the URL with deletion status.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - shortURL: short URL identifier to look up
//
// Returns:
//   - *model.URL: found URL object with deletion status
//   - error: error if URL is not found or database operation fails
func (p *PostgresRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	row := p.db.QueryRowContext(ctx, "select original_url, is_deleted from t_short_url where short_url = $1", shortURL)
	var originalURL string
	var isDeleted bool
	err := row.Scan(&originalURL, &isDeleted)
	if err != nil {
		return nil, err
	}
	var url = model.NewURL(shortURL, originalURL)
	url.IsDeleted = isDeleted
	return url, nil
}

// GetByUserID retrieves all URLs created by a specific user.
// Only returns non-deleted URLs for the user.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: user identifier to look up URLs for
//
// Returns:
//   - []model.URL: slice of URLs created by the user
//   - error: error if database operation fails
func (p *PostgresRepository) GetByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	rows, err := p.db.QueryContext(ctx,
		"select short_url, original_url from t_short_url where user_id = $1 and is_deleted = false",
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

// GetStats retrieves service statistics including total URLs and unique users.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//
// Returns:
//   - int: total number of shortened URLs in the service
//   - int: total number of unique users in the service
//   - error: any error that occurred during statistics retrieval
//
// Note: Access to this method should be restricted to trusted networks.
func (p *PostgresRepository) GetStats(ctx context.Context) (urlsCount int, usersCount int, err error) {
	query := `
		SELECT 
			COALESCE((SELECT COUNT(*) FROM t_short_url WHERE is_deleted = false), 0) as urls_count,
			COALESCE((SELECT COUNT(DISTINCT user_id) FROM t_short_url WHERE is_deleted = false AND user_id IS NOT NULL), 0) as users_count
	`

	err = p.db.QueryRowContext(ctx, query).Scan(&urlsCount, &usersCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get service statistics: %w", err)
	}

	return urlsCount, usersCount, nil
}

// Ping checks the connectivity to PostgreSQL database.
// Used for health checks and connection validation.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//
// Returns:
//   - error: error if database is unreachable
func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Close closes the database connection and releases resources.
// Should be called when the repository is no longer needed.
//
// Returns:
//   - error: error if database closing fails
func (p *PostgresRepository) Close() error {
	return p.db.Close()
}

// getShortURLByOriginalURL retrieves the short URL for a given original URL.
// Internal helper method for conflict resolution.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - originalURL: original URL to look up
//
// Returns:
//   - string: found short URL
//   - error: error if URL is not found or database operation fails
func (p *PostgresRepository) getShortURLByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	row := p.db.QueryRowContext(ctx, "select short_url from t_short_url where original_url = $1", originalURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

// isUniqueViolation checks if an error is a PostgreSQL unique constraint violation.
// Helper function for handling duplicate key errors.
//
// Parameters:
//   - err: error to check
//
// Returns:
//   - bool: true if error is a unique constraint violation
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

// ErrURLConflict represents a URL conflict error when original URL already exists.
// Contains the existing short URL for the original URL.
type ErrURLConflict struct {
	ShortURL string
	Err      string
}

// Error returns the string representation of the URL conflict error.
// Implements the error interface.
//
// Returns:
//   - string: error message
func (err *ErrURLConflict) Error() string {
	return err.Err
}
