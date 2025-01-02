package imageendpoint

import (
	"context"

	"github.com/go-kit/log"
	"github.com/google/uuid"

	"github.com/go-kit/kit/endpoint"
	"github.com/yckao/image-search-demo-go/pkg/models"
	"github.com/yckao/image-search-demo-go/services/image/imageservice"
)

type Endpoints struct {
	logger                 log.Logger
	CreateImageEndpoint    endpoint.Endpoint
	GetImageEndpoint       endpoint.Endpoint
	SearchImageEndpoint    endpoint.Endpoint
	SearchFeedbackEndpoint endpoint.Endpoint
}

func New(svc imageservice.Service, logger log.Logger) Endpoints {
	var searchEndpoint endpoint.Endpoint
	{
		searchEndpoint = MakeSearchImageEndpoint(svc)
	}

	var createImageEndpoint endpoint.Endpoint
	{
		createImageEndpoint = MakeCreateImageEndpoint(svc)
	}

	var getImageEndpoint endpoint.Endpoint
	{
		getImageEndpoint = MakeGetImageEndpoint(svc)
	}

	var searchFeedbackEndpoint endpoint.Endpoint
	{
		searchFeedbackEndpoint = MakeSearchFeedbackEndpoint(svc)
	}

	return Endpoints{
		logger:                 logger,
		CreateImageEndpoint:    createImageEndpoint,
		GetImageEndpoint:       getImageEndpoint,
		SearchImageEndpoint:    searchEndpoint,
		SearchFeedbackEndpoint: searchFeedbackEndpoint,
	}
}

func MakeCreateImageEndpoint(svc imageservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateImageRequest)
		defer req.Closer()

		resp, err := svc.CreateImage(ctx, req.Image)

		return CreateImageResponse{
			V:   resp,
			Err: err,
		}, nil
	}
}

func MakeGetImageEndpoint(svc imageservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetImageRequest)
		resp, err := svc.GetImage(ctx, req.ID)
		return GetImageResponse{
			V:   resp,
			Err: err,
		}, nil
	}
}

func MakeSearchImageEndpoint(svc imageservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SearchImageRequest)
		resp, err := svc.SearchImage(ctx, req.Query)
		return SearchImageResponse{
			V:   resp,
			Err: err,
		}, nil
	}
}

func MakeSearchFeedbackEndpoint(svc imageservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SearchFeedbackRequest)
		resp, err := svc.SearchFeedback(ctx, req.QueryID, req.Rating)
		return SearchFeedbackResponse{
			V:   resp,
			Err: err,
		}, nil
	}
}

var _ imageservice.Service = (*Endpoints)(nil)

func (e *Endpoints) CreateImage(ctx context.Context, image *models.StorageFileStream) (*models.Image, error) {
	resp, err := e.CreateImageEndpoint(ctx, image)
	if err != nil {
		return nil, err
	}
	response := resp.(CreateImageResponse)
	return response.V, response.Err
}

func (e *Endpoints) GetImage(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	resp, err := e.GetImageEndpoint(ctx, id)
	if err != nil {
		return nil, err
	}
	response := resp.(GetImageResponse)
	return response.V, response.Err
}

func (e *Endpoints) SearchImage(ctx context.Context, query string) (*models.SearchWithImage, error) {
	resp, err := e.SearchImageEndpoint(ctx, query)
	if err != nil {
		return nil, err
	}
	response := resp.(SearchImageResponse)
	return response.V, response.Err
}

func (e *Endpoints) SearchFeedback(ctx context.Context, query_id uuid.UUID, rating models.Rating) (*models.SearchFeedbackWithQuery, error) {
	resp, err := e.SearchFeedbackEndpoint(ctx, SearchFeedbackRequest{
		QueryID: query_id,
		Rating:  rating,
	})
	if err != nil {
		return nil, err
	}
	response := resp.(SearchFeedbackResponse)
	return response.V, response.Err
}

var (
	_ endpoint.Failer = CreateImageResponse{}
	_ endpoint.Failer = SearchImageResponse{}
	_ endpoint.Failer = GetImageResponse{}
	_ endpoint.Failer = SearchFeedbackResponse{}
)

type CreateImageRequest struct {
	Image  *models.StorageFileStream
	Closer func() error
}

type CreateImageResponse struct {
	V   *models.Image
	Err error
}

func (r CreateImageResponse) Failed() error {
	return r.Err
}

type GetImageRequest struct {
	ID uuid.UUID
}

type GetImageResponse struct {
	V   *models.Image
	Err error
}

func (r GetImageResponse) Failed() error {
	return r.Err
}

type SearchImageRequest struct {
	Query string
}

type SearchImageResponse struct {
	V   *models.SearchWithImage
	Err error
}

func (r SearchImageResponse) Failed() error {
	return r.Err
}

type SearchFeedbackRequest struct {
	QueryID uuid.UUID
	Rating  models.Rating
}

type SearchFeedbackResponse struct {
	V   *models.SearchFeedbackWithQuery
	Err error
}

func (r SearchFeedbackResponse) Failed() error {
	return r.Err
}
