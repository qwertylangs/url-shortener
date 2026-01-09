package models

import "time"

type URL struct {
    ID        int64     `json:"id"`
    Alias     string    `json:"alias"`
    URL       string    `json:"url"`
    UserID    int64     `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}