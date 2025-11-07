// postgres_test.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bezjen/shortener/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupPostgresRepository(t *testing.T) (*PostgresRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	repo := &PostgresRepository{db: db}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestPostgresRepositorySave(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	tests := []struct {
		name          string
		userID        string
		url           model.URL
		setupMock     func()
		expectedError error
	}{
		{
			name:   "Save successfully",
			userID: "user1",
			url:    *model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
			setupMock: func() {
				mock.ExpectExec("insert into t_short_url").
					WithArgs("qwerty12", "https://practicum.yandex.ru/", "user1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
		{
			name:   "Save with unique violation - URL exists and not deleted",
			userID: "user1",
			url:    *model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
			setupMock: func() {
				// First insert fails with unique violation
				mock.ExpectExec("insert into t_short_url").
					WithArgs("qwerty12", "https://practicum.yandex.ru/", "user1").
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

				// Then query to check if deleted
				rows := sqlmock.NewRows([]string{"short_url", "is_deleted"}).
					AddRow("existing123", false)
				mock.ExpectQuery("select short_url, is_deleted from t_short_url where original_url =").
					WithArgs("https://practicum.yandex.ru/").
					WillReturnRows(rows)
			},
			expectedError: &ErrURLConflict{ShortURL: "existing123", Err: "Original URL already exists"},
		},
		{
			name:   "Save with unique violation - URL exists but deleted",
			userID: "user1",
			url:    *model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
			setupMock: func() {
				// First insert fails with unique violation
				mock.ExpectExec("insert into t_short_url").
					WithArgs("qwerty12", "https://practicum.yandex.ru/", "user1").
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

				// Then query to check if deleted - returns true
				rows := sqlmock.NewRows([]string{"short_url", "is_deleted"}).
					AddRow("existing123", true)
				mock.ExpectQuery("select short_url, is_deleted from t_short_url where original_url =").
					WithArgs("https://practicum.yandex.ru/").
					WillReturnRows(rows)

				// Then update the record
				mock.ExpectExec("update t_short_url set short_url = \\$1, user_id = \\$2, is_deleted = false where original_url =").
					WithArgs("qwerty12", "user1", "https://practicum.yandex.ru/").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := repo.Save(context.TODO(), tt.userID, tt.url)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if conflictErr, ok := tt.expectedError.(*ErrURLConflict); ok {
					if actualConflict, ok := err.(*ErrURLConflict); ok {
						assert.Equal(t, conflictErr.ShortURL, actualConflict.ShortURL)
						assert.Equal(t, conflictErr.Err, actualConflict.Err)
					} else {
						t.Errorf("Expected ErrURLConflict, got %T: %v", err, err)
					}
				} else {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepositorySaveBatch(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	batch := []model.URL{
		*model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
		*model.NewURL("qwerty13", "https://example.com/"),
	}

	mock.ExpectBegin()
	mock.ExpectExec("insert into t_short_url").
		WithArgs("qwerty12", "https://practicum.yandex.ru/", "user1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("insert into t_short_url").
		WithArgs("qwerty13", "https://example.com/", "user1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SaveBatch(context.TODO(), "user1", batch)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryGetByShortURL(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	shortURL := "qwerty12"
	originalURL := "https://practicum.yandex.ru/"
	isDeleted := false

	rows := sqlmock.NewRows([]string{"original_url", "is_deleted"}).
		AddRow(originalURL, isDeleted)

	mock.ExpectQuery("select original_url, is_deleted from t_short_url where short_url =").
		WithArgs(shortURL).
		WillReturnRows(rows)

	result, err := repo.GetByShortURL(context.TODO(), shortURL)
	assert.NoError(t, err)
	assert.Equal(t, shortURL, result.ShortURL)
	assert.Equal(t, originalURL, result.OriginalURL)
	assert.Equal(t, isDeleted, result.IsDeleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryGetByShortURL_NotFound(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	shortURL := "nonexistent"

	mock.ExpectQuery("select original_url, is_deleted from t_short_url where short_url =").
		WithArgs(shortURL).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByShortURL(context.TODO(), shortURL)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryGetByUserID(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	userID := "user1"
	expectedURLs := []model.URL{
		*model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
		*model.NewURL("qwerty13", "https://example.com/"),
	}

	rows := sqlmock.NewRows([]string{"short_url", "original_url"}).
		AddRow("qwerty12", "https://practicum.yandex.ru/").
		AddRow("qwerty13", "https://example.com/")

	mock.ExpectQuery("select short_url, original_url from t_short_url where user_id =").
		WithArgs(userID).
		WillReturnRows(rows)

	urls, err := repo.GetByUserID(context.TODO(), userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedURLs, urls)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryPing(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	mock.ExpectPing()

	err := repo.Ping(context.TODO())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryClose(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	mock.ExpectClose()

	err := repo.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryGetShortURLByOriginalURL(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	originalURL := "https://practicum.yandex.ru/"
	expectedShortURL := "qwerty12"

	rows := sqlmock.NewRows([]string{"short_url"}).AddRow(expectedShortURL)
	mock.ExpectQuery("select short_url from t_short_url where original_url =").
		WithArgs(originalURL).
		WillReturnRows(rows)

	shortURL, err := repo.getShortURLByOriginalURL(context.TODO(), originalURL)
	assert.NoError(t, err)
	assert.Equal(t, expectedShortURL, shortURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryGetShortURLByOriginalURL_NotFound(t *testing.T) {
	repo, mock, cleanup := setupPostgresRepository(t)
	defer cleanup()

	originalURL := "https://nonexistent.com/"

	mock.ExpectQuery("select short_url from t_short_url where original_url =").
		WithArgs(originalURL).
		WillReturnError(sql.ErrNoRows)

	shortURL, err := repo.getShortURLByOriginalURL(context.TODO(), originalURL)
	assert.Error(t, err)
	assert.Equal(t, "", shortURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Unique violation error",
			err:      &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			expected: true,
		},
		{
			name:     "Other PgError",
			err:      &pgconn.PgError{Code: "22000"}, // Some other error code (class 22 - data exception)
			expected: false,
		},
		{
			name:     "Other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUniqueViolation(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrURLConflict_Error(t *testing.T) {
	err := &ErrURLConflict{
		ShortURL: "qwerty12",
		Err:      "Original URL already exists",
	}

	errorMsg := err.Error()
	assert.Equal(t, "Original URL already exists", errorMsg)
}
