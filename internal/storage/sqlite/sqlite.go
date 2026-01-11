package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"

	"url-shortener/internal/models"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", storagePath))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string, userID int64) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias, user_id) VALUES(?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, urlToSave, alias, userID)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRowContext(ctx, alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}

// TODO: implement method
// func (s *Storage) DeleteURL(alias string) error

func(s *Storage) GetUserURLs(ctx context.Context, userID int64) ([]models.URL, error) {
	const op = "storage.sqlite.GetUserURLs"

	stmt, err := s.db.Prepare("SELECT id, url, alias, user_id, created_at, updated_at FROM url WHERE user_id = ? ORDER BY updated_at DESC")
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserURLsNotFound
		}

		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	// сделать пагинацию
	var urls []models.URL = make([]models.URL, 0, 20)
	for rows.Next() {
		var url models.URL
		err := rows.Scan(&url.ID, &url.URL, &url.Alias, &url.UserID, &url.CreatedAt, &url.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: scan row: %w", op, err)
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string, userID int64, isAdmin bool) error {
	const op = "storage.sqlite.DeleteURL"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	
	if !isAdmin {
		var creatorUserID int64
		err := tx.QueryRowContext(ctx, "SELECT user_id FROM url WHERE alias = ?", alias).Scan(&creatorUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storage.ErrURLNotFound
			}
			return fmt.Errorf("%s: query row: %w", op, err)
		}
		if creatorUserID != userID {
			return storage.ErrURLNotOwned
		}
	}
	
	if _, err = tx.ExecContext(ctx, "DELETE FROM url WHERE alias = ?", alias); err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
        return fmt.Errorf("%s: commit transaction: %w", op, err)
    }

	return nil
}