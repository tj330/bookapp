package grpc

import (
	"context"

	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/internal/grpcutil"
	"github.com/tj330/bookapp/metadata/pkg/model"
	"github.com/tj330/bookapp/pkg/discovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type Gateway struct {
	registry discovery.Registry
	creds    credentials.TransportCredentials
}

func New(registry discovery.Registry, creds credentials.TransportCredentials) *Gateway {
	return &Gateway{registry: registry, creds: creds}
	//return &Gateway{registry: registry}
}

func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	//conn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry, g.creds)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewMetadataServiceClient(conn)
	var resp *gen.GetMetadataResponse
	const maxRetries = 5

	for range maxRetries {
		resp, err = client.GetMetadata(ctx, &gen.GetMetadataRequest{BookId: id})
		if err != nil {
			if shouldRetry(err) {
				continue
			}
			return nil, err
		}
		return model.MetadataFromProto(resp.Metadata), nil
	}
	return nil, err
}

func shouldRetry(err error) bool {
	e, ok := status.FromError(err)
	if !ok {
		return false
	}
	return e.Code() == codes.DeadlineExceeded || e.Code() == codes.ResourceExhausted || e.Code() == codes.Unavailable
}
