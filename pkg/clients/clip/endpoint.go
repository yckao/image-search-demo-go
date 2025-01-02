package clip

import (
	context "context"
	"io"

	"github.com/go-kit/kit/endpoint"
	"github.com/yckao/image-search-demo-go/pkg/models"
)

type Endpoints struct {
	ImageEmbeddingEndpoint endpoint.Endpoint
	TextEmbeddingEndpoint  endpoint.Endpoint
}

func (e *Endpoints) ImageEmbedding(ctx context.Context, image io.Reader) (*models.Embedding, error) {
	resp, err := e.ImageEmbeddingEndpoint(ctx, &ImageEmbeddingRequest{Image: image})
	if err != nil {
		return nil, err
	}
	response := resp.(ImageEmbeddingResponse)
	return response.V, response.Err
}

func (e *Endpoints) TextEmbedding(ctx context.Context, text string) (*models.Embedding, error) {
	resp, err := e.TextEmbeddingEndpoint(ctx, &TextEmbeddingRequest{Text: text})
	if err != nil {
		return nil, err
	}
	response := resp.(*TextEmbeddingResponse)
	return response.V, response.Err
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = ImageEmbeddingResponse{}
	_ endpoint.Failer = TextEmbeddingResponse{}
)

type ImageEmbeddingRequest struct {
	Image io.Reader
}

type ImageEmbeddingResponse struct {
	V   *models.Embedding
	Err error
}

func (r ImageEmbeddingResponse) Failed() error { return r.Err }

type TextEmbeddingRequest struct {
	Text string
}

type TextEmbeddingResponse struct {
	V   *models.Embedding
	Err error
}

func (r TextEmbeddingResponse) Failed() error { return r.Err }
