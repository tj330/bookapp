package main

import (
	"context"
	"log"
	"net"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	booktest "github.com/tj330/bookapp/book/pkg/testutil"
	"github.com/tj330/bookapp/gen"
	metadatatest "github.com/tj330/bookapp/metadata/pkg/testutil"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/memory"
	ratingtest "github.com/tj330/bookapp/rating/pkg/testutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	metadataServiceName    = "metadata"
	ratingServiceName      = "rating"
	bookServiceName        = "book"
	metadataServiceAddress = "localhost:8081"
	ratingServiceAddress   = "localhost:8082"
	bookServiceAddress     = "localhost:8083"
)

func main() {
	log.Println("Starting the integration test")

	ctx := context.Background()
	registry := memory.New()

	log.Println("Setting up service handlers and clients")

	metadataSrv := startMetadataService(ctx, registry)
	defer metadataSrv.GracefulStop()
	ratingSrv := startRatingService(ctx, registry)
	defer ratingSrv.GracefulStop()
	bookSrv := startBookService(ctx, registry)
	defer bookSrv.GracefulStop()

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	metadataConn, err := grpc.NewClient(metadataServiceAddress, opts)
	if err != nil {
		panic(err)
	}
	defer metadataConn.Close()
	metadataClient := gen.NewMetadataServiceClient(metadataConn)

	ratingConn, err := grpc.NewClient(ratingServiceAddress, opts)
	if err != nil {
		panic(err)
	}
	defer ratingConn.Close()
	ratingClient := gen.NewRatingServiceClient(ratingConn)

	bookConn, err := grpc.NewClient(bookServiceAddress, opts)
	if err != nil {
		panic(err)
	}
	defer bookConn.Close()
	bookClient := gen.NewBookServiceClient(bookConn)

	log.Println("Saving test metada via metadata service")

	m := &gen.Metadata{
		Id:          "the-book",
		Title:       "The Book",
		Description: "The one and only book",
		Author:      "Mr. TJ",
		Isbn:        "123456789101",
	}

	if _, err := metadataClient.PutMetadata(ctx, &gen.PutMetadataRequest{Metadata: m}); err != nil {
		log.Fatalf("put metadata: %v", err)
	}

	log.Println("retrieving test metadata via metadata service")

	getMetadataResp, err := metadataClient.GetMetadata(ctx, &gen.GetMetadataRequest{BookId: m.Id})
	if err != nil {
		log.Fatalf("get metadata: %v", err)
	}

	if diff := cmp.Diff(getMetadataResp.Metadata, m, cmpopts.IgnoreUnexported(gen.Metadata{})); diff != "" {
		log.Fatalf("get metadata after put mismatch: %v", err)
	}

	log.Println("Getting book details via book service")

	wantBookDetails := &gen.BookDetails{Metadata: m}

	getBookDetailsResp, err := bookClient.GetBookDetails(ctx, &gen.GetBookDetailsRequest{BookId: m.Id})
	if err != nil {
		log.Fatalf("get book details: %v", err)
	}

	if diff := cmp.Diff(getBookDetailsResp.BookDetails, wantBookDetails, cmpopts.IgnoreUnexported(gen.BookDetails{}, gen.Metadata{})); diff != "" {
		log.Fatalf("get book details after put mismatch: %v", err)
	}

	log.Println("Saving first rating via rating service")

	const userID = "user0"
	const recordTypeBook = "book"
	firstRating := int32(5)

	if _, err := ratingClient.PutRating(ctx, &gen.PutRatingRequest{UserId: userID, RecordId: m.Id, RecordType: recordTypeBook, RatingValue: firstRating}); err != nil {
		log.Fatalf("put rating: %v", err)
	}

	log.Println("Retrieving initial aggregated rating via rating service")

	getAggregatedRatingResp, err := ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: m.Id, RecordType: recordTypeBook})
	if err != nil {
		log.Fatalf("get aggregated rating: %v", err)
	}

	if got, want := getAggregatedRatingResp.RatingValue, float64(5); got != want {
		log.Fatalf("rating mismatch. got %v want %v", got, want)
	}

	log.Println("Saving second rating via rating service")

	secondRating := int32(1)

	if _, err := ratingClient.PutRating(ctx, &gen.PutRatingRequest{UserId: userID, RecordId: m.Id, RecordType: recordTypeBook, RatingValue: secondRating}); err != nil {
		log.Fatalf("put rating: %v", err)
	}

	log.Println("Saving new aggregated rating via rating service")

	getAggregatedRatingResp, err = ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: m.Id, RecordType: recordTypeBook})
	if err != nil {
		log.Fatalf("get aggregated rating: %v", err)
	}
	wantRating := float64((firstRating + secondRating) / 2)
	if got, want := getAggregatedRatingResp.RatingValue, wantRating; got != want {
		log.Fatalf("rating mismatch. got %v want %v", got, want)
	}

	log.Println("getting update book details via book service")

	getBookDetailsResp, err = bookClient.GetBookDetails(ctx, &gen.GetBookDetailsRequest{BookId: m.Id})
	if err != nil {
		log.Fatalf("get book details: %v", err)
	}

	wantBookDetails.Rating = wantRating

	if diff := cmp.Diff(getBookDetailsResp.BookDetails, wantBookDetails, cmpopts.IgnoreUnexported(gen.BookDetails{}, gen.Metadata{})); diff != "" {
		log.Fatalf("get movie details after update mismatch: %v", err)
	}
	log.Println("Integration test execution successful")
}

func startMetadataService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting metadata service on " + metadataServiceAddress)
	h := metadatatest.NewTestMetadataGRPCServer()
	l, err := net.Listen("tcp", metadataServiceAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterMetadataServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()

	id := discovery.GenerateInstanceID(metadataServiceName)
	if err := registry.Register(ctx, id, metadataServiceName, metadataServiceAddress); err != nil {
		panic(err)
	}
	return srv
}

func startRatingService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting rating service on " + ratingServiceAddress)
	h := ratingtest.NewTestRatingGRPCServer()
	l, err := net.Listen("tcp", ratingServiceAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterRatingServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()

	id := discovery.GenerateInstanceID(ratingServiceName)
	if err := registry.Register(ctx, id, ratingServiceName, ratingServiceAddress); err != nil {
		panic(err)
	}
	return srv
}

func startBookService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting book service on " + bookServiceAddress)
	h := booktest.NewTestBookGRPCServer(registry)
	l, err := net.Listen("tcp", bookServiceAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterBookServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()

	id := discovery.GenerateInstanceID(bookServiceName)
	if err := registry.Register(ctx, id, bookServiceName, bookServiceAddress); err != nil {
		panic(err)
	}
	return srv
}
