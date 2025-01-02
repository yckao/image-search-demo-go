package storageservice

import (
	"context"

	"github.com/yckao/image-search-demo-go/pkg/models"
)

type Service interface {
	Upload(ctx context.Context, stream *models.StorageFileStream) (*models.StorageFile, error)
	Download(ctx context.Context, file *models.StorageFile) (*models.StorageFileStream, error)
	FormatURL(ctx context.Context, file *models.StorageFile) (string, error)
}
