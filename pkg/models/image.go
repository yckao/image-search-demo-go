package models

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID              uuid.UUID `json:"id"`
	StorageProvider string    `json:"storage_provider"`
	StorageKey      string    `json:"storage_key"`
	CreatedAt       time.Time `json:"created_at"`
	URL             string    `json:"url"`
}

type Search struct {
	ID        uuid.UUID `json:"id"`
	ModelName string    `json:"model_name"`
	QueryText string    `json:"query_text"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchWithImage struct {
	Search
	Image Image `json:"image"`
}

type Rating string

const (
	RatingPositive Rating = "POSITIVE"
	RatingNegative        = "NEGATIVE"
)

type SearchFeedback struct {
	ID        uuid.UUID `json:"id"`
	Rating    Rating    `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchFeedbackWithQuery struct {
	SearchFeedback
	Query Search `json:"query"`
}
