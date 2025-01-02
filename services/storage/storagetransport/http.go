package storagetransport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/yckao/image-search-demo-go/pkg/errortypes"
	"github.com/yckao/image-search-demo-go/services/storage/storageendpoint"
)

func NewHTTPHandler(svc storageendpoint.Endpoints, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errortypes.ErrorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	m := http.NewServeMux()

	m.Handle("GET /storage/{provider}/files/{key...}", httptransport.NewServer(
		svc.DownloadEndpoint,
		decodeDownloadRequest,
		encodeResponse,
		options...,
	))

	return m
}

func decodeDownloadRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return storageendpoint.DownloadRequest{
		Provider: r.PathValue("provider"),
		Key:      r.PathValue("key"),
	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(storageendpoint.DownloadResponse)

	if resp.Err != nil {
		errortypes.ErrorEncoder(ctx, resp.Err, w)
		return nil
	}

	w.Header().Set("Content-Type", resp.V.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(resp.V.ContentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.V.Filename))

	_, err := io.Copy(w, resp.V.Reader)

	return err
}
