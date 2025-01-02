package storageendpoint

import (
	"context"
	"io"

	"github.com/go-kit/log"

	"github.com/go-kit/kit/endpoint"
	"github.com/yckao/image-search-demo-go/pkg/models"
	"github.com/yckao/image-search-demo-go/services/storage/storageservice"
)

type Endpoints struct {
	logger            log.Logger
	UploadEndpoint    endpoint.Endpoint
	DownloadEndpoint  endpoint.Endpoint
	FormatURLEndpoint endpoint.Endpoint
}

func New(svc storageservice.Service, logger log.Logger) Endpoints {
	var uploadEndpoint endpoint.Endpoint
	{
		uploadEndpoint = MakeUploadEndpoint(svc)
	}

	var downloadEndpoint endpoint.Endpoint
	{
		downloadEndpoint = MakeDownloadEndpoint(svc)
	}

	var formatURLEndpoint endpoint.Endpoint
	{
		formatURLEndpoint = MakeFormatURLEndpoint(svc)
	}

	return Endpoints{
		logger:            logger,
		UploadEndpoint:    uploadEndpoint,
		DownloadEndpoint:  downloadEndpoint,
		FormatURLEndpoint: formatURLEndpoint,
	}
}

func MakeDownloadEndpoint(svc storageservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DownloadRequest)
		resp, err := svc.Download(ctx, &models.StorageFile{
			Provider: req.Provider,
			Key:      req.Key,
		})
		return DownloadResponse{
			V:   resp,
			Err: err,
		}, nil
	}
}

func MakeUploadEndpoint(svc storageservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UploadRequest)

		resp, err := svc.Upload(ctx, &models.StorageFileStream{
			Reader:        req.Reader,
			ContentType:   req.ContentType,
			ContentLength: req.ContentLength,
			Filename:      req.Filename,
		})

		return UploadResponse{
			V:   resp,
			Err: err,
		}, nil
	}
}

func MakeFormatURLEndpoint(svc storageservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FormatURLRequest)
		resp, err := svc.FormatURL(ctx, req.File)
		return FormatURLResponse{
			V:   resp,
			Err: err,
		}, nil
	}
}

var _ storageservice.Service = (*Endpoints)(nil)

func (e *Endpoints) Upload(ctx context.Context, stream *models.StorageFileStream) (*models.StorageFile, error) {
	resp, err := e.UploadEndpoint(ctx, stream)
	if err != nil {
		return nil, err
	}
	response := resp.(UploadResponse)
	return response.V, response.Err
}

func (e *Endpoints) Download(ctx context.Context, file *models.StorageFile) (*models.StorageFileStream, error) {
	resp, err := e.DownloadEndpoint(ctx, file)
	if err != nil {
		return nil, err
	}
	response := resp.(DownloadResponse)
	return response.V, response.Err
}

func (e *Endpoints) FormatURL(ctx context.Context, file *models.StorageFile) (string, error) {
	resp, err := e.FormatURLEndpoint(ctx, file)
	if err != nil {
		return "", err
	}
	response := resp.(FormatURLResponse)
	return response.V, response.Err
}

var (
	_ endpoint.Failer = UploadResponse{}
	_ endpoint.Failer = DownloadResponse{}
	_ endpoint.Failer = FormatURLResponse{}
)

type UploadRequest struct {
	Reader        io.Reader
	Filename      string
	ContentType   string
	ContentLength int64
}

type UploadResponse struct {
	V   *models.StorageFile
	Err error
}

func (r UploadResponse) Failed() error {
	return r.Err
}

type DownloadRequest struct {
	Provider string `json:"provider"`
	Key      string `json:"key"`
}

type DownloadResponse struct {
	V   *models.StorageFileStream
	Err error
}

func (r DownloadResponse) Failed() error {
	return r.Err
}

type FormatURLRequest struct {
	BaseURL string
	File    *models.StorageFile
}

type FormatURLResponse struct {
	V   string
	Err error
}

func (r FormatURLResponse) Failed() error {
	return r.Err
}
