package testutil

import (
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/rating/internal/controller/rating"
	grpchandler "github.com/tj330/bookapp/rating/internal/handler/grpc"
	"github.com/tj330/bookapp/rating/internal/repository/memory"
)

func NewTestRatingGRPCServer() gen.RatingServiceServer {
	r := memory.New()
	ctrl := rating.New(r, nil)
	return grpchandler.New(ctrl)
}
