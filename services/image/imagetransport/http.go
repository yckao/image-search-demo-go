package imagetransport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/google/uuid"

	"github.com/yckao/image-search-demo-go/pkg/errortypes"
	"github.com/yckao/image-search-demo-go/pkg/models"
	"github.com/yckao/image-search-demo-go/services/image/imageendpoint"
)

func NewHTTPHandler(svc imageendpoint.Endpoints, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errortypes.ErrorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	m := http.NewServeMux()

	m.Handle("POST /images", httptransport.NewServer(
		svc.CreateImageEndpoint,
		decodeCreateImageRequest,
		encodeCreateImageResponse,
		options...,
	))

	m.Handle("GET /images/{id}", httptransport.NewServer(
		svc.GetImageEndpoint,
		decodeGetImageRequest,
		encodeGetImageResponse,
		options...,
	))

	m.Handle("GET /images", httptransport.NewServer(
		svc.SearchImageEndpoint,
		decodeSearchImageRequest,
		encodeSearchImageResponse,
		options...,
	))

	m.Handle("POST /images/{id}/feedback", httptransport.NewServer(
		svc.SearchFeedbackEndpoint,
		decodeSearchFeedbackRequest,
		encodeSearchFeedbackResponse,
		options...,
	))

	return m
}

func decodeCreateImageRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	const maxFileSize = 10 << 20

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get form file: %w", err)
	}

	return imageendpoint.CreateImageRequest{
		Image: &models.StorageFileStream{
			Reader:        file,
			Filename:      header.Filename,
			ContentType:   header.Header.Get("Content-Type"),
			ContentLength: header.Size,
		},
		Closer: func() error {
			return file.Close()
		},
	}, nil
}

func encodeCreateImageResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(imageendpoint.CreateImageResponse)

	if resp.Err != nil {
		errortypes.ErrorEncoder(ctx, resp.Err, w)
		return nil
	}

	return json.NewEncoder(w).Encode(resp.V)
}

func decodeGetImageRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse id: %w", err)
	}

	return imageendpoint.GetImageRequest{
		ID: id,
	}, nil
}

func encodeGetImageResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(imageendpoint.GetImageResponse)

	if resp.Err != nil {
		errortypes.ErrorEncoder(ctx, resp.Err, w)
		return nil
	}

	return json.NewEncoder(w).Encode(resp.V)
}

func decodeSearchImageRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	query := r.URL.Query().Get("query")

	return imageendpoint.SearchImageRequest{
		Query: query,
	}, nil
}

func encodeSearchImageResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(imageendpoint.SearchImageResponse)

	if resp.Err != nil {
		errortypes.ErrorEncoder(ctx, resp.Err, w)
		return nil
	}

	return json.NewEncoder(w).Encode(resp.V)
}

func decodeSearchFeedbackRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse id: %w", err)
	}

	rating := models.Rating(r.FormValue("rating"))

	return imageendpoint.SearchFeedbackRequest{
		QueryID: id,
		Rating:  rating,
	}, nil
}

func encodeSearchFeedbackResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(imageendpoint.SearchFeedbackResponse)

	if resp.Err != nil {
		errortypes.ErrorEncoder(ctx, resp.Err, w)
		return nil
	}

	return json.NewEncoder(w).Encode(resp.V)
}
