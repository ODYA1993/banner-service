package banner

import (
	"time"
)

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Feature struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Banner struct {
	ID        int       `json:"id"`
	Title     string    `json:"title" validate:"required"`
	Text      string    `json:"text" validate:"required"`
	URL       string    `json:"url" validate:"required"`
	IsActive  bool      `json:"is_active"`
	FeatureID Feature   `json:"feature_id" validate:"required"`
	Tags      []Tag     `json:"tags" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
