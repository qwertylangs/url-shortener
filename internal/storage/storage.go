package storage

import (
	"context"
	"errors"
	"url-shortener/internal/models"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
	ErrUserURLsNotFound = errors.New("user urls not found")
)

// Storage represents the storage interface for URL operations


type Storage interface {
	SaveURL(ctx context.Context, urlToSave string, alias string, userID string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)
	// TODO: add DeleteURL method
}
