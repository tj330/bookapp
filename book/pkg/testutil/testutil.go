package testutil

import (
	"github.com/tj330/bookapp/book/internal/controller/book"
	metadatagateway "github.com/tj330/bookapp/book/internal/gateway/metadata/grpc"
	ratinggateway "github.com/tj330/bookapp/book/internal/gateway/rating/grpc"
	grpchandler "github.com/tj330/bookapp/book/internal/handler/grpc"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/pkg/discovery"
	"google.golang.org/grpc/credentials/insecure"
)

func NewTestBookGRPCServer(registry discovery.Registry) gen.BookServiceServer {
	metadataGateway := metadatagateway.New(registry, insecure.NewCredentials())
	ratingGateway := ratinggateway.New(registry)
	ctrl := book.New(ratingGateway, metadataGateway)
	return grpchandler.New(ctrl)
}
