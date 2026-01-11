package storage

import (
	"context"
	"errors"
	"url-shortener/internal/models"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLNotOwned = errors.New("url not owned")
	ErrURLExists   = errors.New("url exists")
	ErrUserURLsNotFound = errors.New("user urls not found")
)

// Storage represents the storage interface for URL operations


type Storage interface {
	SaveURL(ctx context.Context, urlToSave string, alias string, userID int64) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	GetUserURLs(ctx context.Context, userID int64) ([]models.URL, error)
	DeleteURL(ctx context.Context, alias string, userID int64, isAdmin bool) error
}
