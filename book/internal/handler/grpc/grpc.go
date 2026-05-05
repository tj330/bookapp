package grpc

import (
	"context"
	"errors"

	"github.com/tj330/bookapp/book/internal/controller/book"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/metadata/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedBookServiceServer
	ctrl *book.Controller
}

func New(ctrl *book.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

func (h *Handler) GetBookDetails(ctx context.Context, req *gen.GetBookDetailsRequest) (*gen.GetBookDetailsResponse, error) {
	if req == nil || req.BookId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}
	m, err := h.ctrl.Get(ctx, req.BookId)
	if err != nil && errors.Is(err, book.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.GetBookDetailsResponse{
		BookDetails: &gen.BookDetails{
			Metadata: model.MetadataToProto(&m.Metadata),
			Rating:   *m.Rating,
		},
	}, nil
}
