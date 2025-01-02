package imagerepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yckao/image-search-demo-go/pkg/models"
	"github.com/yckao/image-search-demo-go/services/image/imagemodel"
)

type Repository interface {
	CreateImage(ctx context.Context, image *models.Image, embedding *imagemodel.ImageEmbedding) (*models.Image, error)
	GetImage(ctx context.Context, id uuid.UUID) (*models.Image, error)
	CreateSearchQuery(ctx context.Context, searchQuery *imagemodel.SearchQuery) (*models.SearchWithImage, error)
	GetSearchQuery(ctx context.Context, id uuid.UUID) (*models.SearchWithImage, error)
	CreateSearchFeedback(ctx context.Context, feedback *models.SearchFeedbackWithQuery) (*models.SearchFeedbackWithQuery, error)
}
