package grpc

import (
	"context"

	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/internal/grpcutil"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/rating/pkg/model"
	"google.golang.org/grpc/credentials"
)

// Gateway manages traffic by discovering services
// and establishing secure connections to them.
type Gateway struct {
	registry discovery.Registry
	creds    credentials.TransportCredentials
}

// New returns a new Gateway with the provided service discovery registry and
// transport credentials.
func New(registry discovery.Registry, creds credentials.TransportCredentials) *Gateway {
	return &Gateway{registry: registry, creds: creds}
}

// GetAggregatedRating returns the aggregated book rating using the recordID(book id) and recordType
// by communicating with the rating service
func (g *Gateway) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	// conn, err := grpc.NewClient(
	// 	"localhost:8082",
	// 	grpc.WithTransportCredentials(
	// 		insecure.NewCredentials(),
	// 	),
	// )
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry, g.creds)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := gen.NewRatingServiceClient(conn)
	resp, err := client.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: string(recordID), RecordType: string(recordType)})
	if err != nil {
		return 0, nil
	}
	return resp.RatingValue, nil
}
