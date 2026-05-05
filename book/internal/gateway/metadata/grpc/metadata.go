package grpc

import (
	"context"

	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/internal/grpcutil"
	"github.com/tj330/bookapp/metadata/pkg/model"
	"github.com/tj330/bookapp/pkg/discovery"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	registry discovery.Registry
	creds    credentials.TransportCredentials
}

func New(registry discovery.Registry, creds credentials.TransportCredentials) *Gateway {
	return &Gateway{registry: registry, creds: creds}
}

func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry, insecure.NewCredentials())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &gen.GetMetadataRequest{BookId: id})
	if err != nil {
		return nil, err
	}
	return model.MetadataFromProto(resp.Metadata), nil
}
