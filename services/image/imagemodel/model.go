package imagemodel

import (
	"time"

	"github.com/google/uuid"
	"github.com/yckao/image-search-demo-go/pkg/models"
)

type ImageEmbedding struct {
	ID        uuid.UUID    `json:"id"`
	Image     models.Image `json:"image"`
	ModelName string       `json:"model_name"`
	Embedding []float32    `json:"embedding"`
	CreatedAt time.Time    `json:"created_at"`
}

type ImageWithEmbedding struct {
	models.Image
	Embedding ImageEmbedding `json:"embedding"`
}

type SearchQuery struct {
	models.Search
	Embedding []float32 `json:"embedding"`
}
