package storageservice

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/go-kit/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/yckao/image-search-demo-go/pkg/errortypes"
	"github.com/yckao/image-search-demo-go/pkg/models"
)

type s3Service struct {
	logger log.Logger
	config S3ServiceConfig
	client *s3.Client
}

type S3ServiceConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	BaseURL   string
	URLFormat string
}

func NewS3Service(logger log.Logger, config S3ServiceConfig) Service {
	client := s3.New(s3.Options{
		Region:           "asia-northeast1",
		EndpointResolver: s3.EndpointResolverFromURL(config.Endpoint),
		Credentials:      credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, ""),
		UsePathStyle:     true,
	})

	return &s3Service{
		logger: logger,
		config: config,
		client: client,
	}
}

func (s *s3Service) Upload(ctx context.Context, stream *models.StorageFileStream) (*models.StorageFile, error) {
	key := fmt.Sprintf("images/%s/%s", nanoid.Must(10), stream.Filename)

	if _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.config.Bucket),
		Key:           aws.String(key),
		ContentLength: aws.Int64(stream.ContentLength),
		ContentType:   aws.String(stream.ContentType),
		Body:          stream.Reader,
	}); err != nil {
		return nil, err
	}

	return &models.StorageFile{
		Provider: "s3",
		Key:      key,
	}, nil
}

func (s *s3Service) Download(ctx context.Context, file *models.StorageFile) (*models.StorageFileStream, error) {
	headers, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(file.Key),
	})

	if err != nil {
		var notFoundErr *types.NotFound
		if errors.As(err, &notFoundErr) {
			return nil, errortypes.NewErrStorageFileNotFound("s3", file.Key)
		}

		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, errortypes.NewErrStorageFileNotFound("s3", file.Key)
		}

		return nil, err
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(file.Key),
	})

	return &models.StorageFileStream{
		Reader:        result.Body,
		ContentType:   *headers.ContentType,
		ContentLength: *headers.ContentLength,
		Filename:      path.Base(file.Key),
	}, nil
}

func (s *s3Service) FormatURL(ctx context.Context, file *models.StorageFile) (string, error) {
	return fmt.Sprintf(s.config.URLFormat, s.config.BaseURL, file.Key), nil
}
