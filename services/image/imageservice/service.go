package imageservice

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/yckao/image-search-demo-go/pkg/clients/clip"
	"github.com/yckao/image-search-demo-go/pkg/models"
	"github.com/yckao/image-search-demo-go/services/image/imagemodel"
	"github.com/yckao/image-search-demo-go/services/image/imagerepository"
	"github.com/yckao/image-search-demo-go/services/storage/storageservice"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	CreateImage(ctx context.Context, image *models.StorageFileStream) (*models.Image, error)
	GetImage(ctx context.Context, id uuid.UUID) (*models.Image, error)
	SearchImage(ctx context.Context, query string) (*models.SearchWithImage, error)
	SearchFeedback(ctx context.Context, query_id uuid.UUID, rating models.Rating) (*models.SearchFeedbackWithQuery, error)
}

type imageService struct {
	logger          log.Logger
	clipService     clip.Service
	storageService  storageservice.Service
	imageRepository imagerepository.Repository
}

func New(logger log.Logger, clipService clip.Service, storageService storageservice.Service, imageRepository imagerepository.Repository) Service {
	return &imageService{
		logger:          logger,
		clipService:     clipService,
		storageService:  storageService,
		imageRepository: imageRepository,
	}
}

func (s *imageService) CreateImage(ctx context.Context, stream *models.StorageFileStream) (*models.Image, error) {
	imageBytes, err := io.ReadAll(stream.Reader)
	if err != nil {
		return nil, err
	}

	var embedding *models.Embedding
	var storageFile *models.StorageFile

	errGroup, errCtx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		e, err := s.clipService.ImageEmbedding(errCtx, bytes.NewReader(imageBytes))
		if err != nil {
			return err
		}
		embedding = e
		return nil
	})

	errGroup.Go(func() error {
		f, err := s.storageService.Upload(errCtx, &models.StorageFileStream{
			Reader:        bytes.NewReader(imageBytes),
			Filename:      stream.Filename,
			ContentType:   stream.ContentType,
			ContentLength: stream.ContentLength,
		})
		if err != nil {
			return err
		}
		storageFile = f
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}

	url, err := s.storageService.FormatURL(ctx, storageFile)
	if err != nil {
		return nil, err
	}

	image, err := s.imageRepository.CreateImage(ctx, &models.Image{
		StorageProvider: storageFile.Provider,
		StorageKey:      storageFile.Key,
		CreatedAt:       time.Now(),
		URL:             url,
	}, &imagemodel.ImageEmbedding{
		ModelName: embedding.Model,
		Embedding: embedding.Embedding,
	})
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *imageService) GetImage(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	img, err := s.imageRepository.GetImage(ctx, id)
	if err != nil {
		return nil, err
	}

	if img.URL, err = s.storageService.FormatURL(ctx, &models.StorageFile{
		Provider: img.StorageProvider,
		Key:      img.StorageKey,
	}); err != nil {
		return nil, err
	}

	return img, nil
}

func (s *imageService) SearchImage(ctx context.Context, query string) (*models.SearchWithImage, error) {
	embedding, err := s.clipService.TextEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	searchWithImage, err := s.imageRepository.CreateSearchQuery(ctx, &imagemodel.SearchQuery{
		Search: models.Search{
			ModelName: embedding.Model,
			QueryText: query,
		},
		Embedding: embedding.Embedding,
	})
	if err != nil {
		return nil, err
	}

	if searchWithImage.Image.URL, err = s.storageService.FormatURL(ctx, &models.StorageFile{
		Provider: searchWithImage.Image.StorageProvider,
		Key:      searchWithImage.Image.StorageKey,
	}); err != nil {
		return nil, err
	}

	return searchWithImage, nil
}

func (s *imageService) SearchFeedback(ctx context.Context, query_id uuid.UUID, rating models.Rating) (*models.SearchFeedbackWithQuery, error) {
	return s.imageRepository.CreateSearchFeedback(ctx, &models.SearchFeedbackWithQuery{
		SearchFeedback: models.SearchFeedback{
			Rating: rating,
		},
		Query: models.Search{
			ID: query_id,
		},
	})
}
