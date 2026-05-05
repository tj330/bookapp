package testutil

import (
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/metadata/internal/controller/metadata"
	grpchandler "github.com/tj330/bookapp/metadata/internal/handler/grpc"
	"github.com/tj330/bookapp/metadata/internal/repository/memory"
)

func NewTestMetadataGRPCServer() gen.MetadataServiceServer {
	r := memory.New()
	ctrl := metadata.New(r)
	return grpchandler.New(ctrl)
}
