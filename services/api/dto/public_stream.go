package dto

import (
	"time"
)

// PublicStreamDTO 公開串流DTO
type PublicStreamDTO struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
