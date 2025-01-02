package clip

import (
	"context"
	"io"

	"github.com/yckao/image-search-demo-go/pkg/models"
)

type Service interface {
	ImageEmbedding(ctx context.Context, image io.Reader) (*models.Embedding, error)
	TextEmbedding(ctx context.Context, text string) (*models.Embedding, error)
}
