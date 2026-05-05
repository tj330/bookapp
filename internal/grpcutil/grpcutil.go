package grpcutil

import (
	"context"
	"math/rand"

	"github.com/tj330/bookapp/pkg/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func ServiceConnection(ctx context.Context, serviceName string, registry discovery.Registry, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	addrs, err := registry.ServiceAddresses(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	return grpc.NewClient(addrs[rand.Intn(len(addrs))], grpc.WithTransportCredentials(creds))
}
