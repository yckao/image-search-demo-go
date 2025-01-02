package clip

import (
	"context"
	"io"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/yckao/image-search-demo-go/pkg/models"
	"google.golang.org/grpc"
)

const (
	maxChunkSize = 64 * 1024 // 64KB chunks
)

type client struct {
	CLIPServiceClient
	conn *grpc.ClientConn
}

func NewGRPCClient(conn *grpc.ClientConn) (Service, error) {
	var options []grpctransport.ClientOption

	var imageEmbeddingEndpoint endpoint.Endpoint
	{
		imageEmbeddingEndpoint = NewStreamClient(conn).Endpoint()
	}

	var textEmbeddingEndpoint endpoint.Endpoint
	{
		textEmbeddingEndpoint = grpctransport.NewClient(
			conn,
			"clip.CLIPService",
			"TextEmbedding",
			encodeGRPCTextEmbeddingRequest,
			decodeGRPCTextEmbeddingResponse,
			EmbeddingResponse{},
			options...,
		).Endpoint()
	}

	return &Endpoints{
		ImageEmbeddingEndpoint: imageEmbeddingEndpoint,
		TextEmbeddingEndpoint:  textEmbeddingEndpoint,
	}, nil
}

func encodeGRPCTextEmbeddingRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req := request.(*TextEmbeddingRequest)

	return &Text{Text: req.Text}, nil
}

func decodeGRPCTextEmbeddingResponse(ctx context.Context, response interface{}) (interface{}, error) {
	res := response.(*EmbeddingResponse)

	return &TextEmbeddingResponse{
		V: &models.Embedding{
			Model:     res.GetModelName(),
			Embedding: res.GetEmbedding(),
		},
	}, nil
}

func NewStreamClient(conn *grpc.ClientConn) *StreamClient {
	return &StreamClient{
		client: NewCLIPServiceClient(conn),
	}
}

type StreamClient struct {
	client CLIPServiceClient
}

func (c *StreamClient) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*ImageEmbeddingRequest)

		stream, err := c.client.ImageEmbedding(ctx)
		if err != nil {
			return nil, err
		}

		buffer := make([]byte, maxChunkSize)

		for {
			n, err := req.Image.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			chunk := buffer[:n]
			req := &ImageChunk{
				Data: chunk,
			}

			if err := stream.Send(req); err != nil {
				return nil, err
			}
		}
		resp, err := stream.CloseAndRecv()
		if err != nil {
			return nil, err
		}

		return ImageEmbeddingResponse{
			V: &models.Embedding{
				Model:     resp.GetModelName(),
				Embedding: resp.GetEmbedding(),
			},
		}, nil
	}
}
