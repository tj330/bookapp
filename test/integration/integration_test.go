package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/pkg/discovery/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcmetadata "google.golang.org/grpc/metadata"
)

func TestEndtoEndFlow(t *testing.T) {
	ctx := context.Background()
	registry := memory.New()

	// Environment setup
	startMetadataService(t, ctx, registry)
	startRatingService(t, ctx, registry)
	startBookService(t, ctx, registry)
	startAuthService(t, ctx, registry)

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())

	// Client connections
	metadataConn, err := grpc.NewClient(metadataServiceAddress, opts)
	if err != nil {
		t.Fatalf("failed to create metadata connection: %v", err)
	}
	defer metadataConn.Close()
	metadataClient := gen.NewMetadataServiceClient(metadataConn)

	ratingConn, err := grpc.NewClient(ratingServiceAddress, opts)
	if err != nil {
		t.Fatalf("failed to create rating connection: %v", err)
	}
	defer ratingConn.Close()
	ratingClient := gen.NewRatingServiceClient(ratingConn)

	bookConn, err := grpc.NewClient(bookServiceAddress, opts)
	if err != nil {
		t.Fatalf("failed to create book connection: %v", err)
	}
	defer bookConn.Close()
	bookClient := gen.NewBookServiceClient(bookConn)

	authConn, err := grpc.NewClient(authServiceAddress, opts)
	if err != nil {
		t.Fatalf("failed to create auth connection: %v", err)
	}
	defer authConn.Close()
	authClient := gen.NewAuthServiceClient(authConn)

	// Shared state / authentication
	tokenResp, err := authClient.GetToken(ctx, &gen.GetTokenRequest{
		Username: "user0",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("get auth token failed: %v", err)
	}

	authCtx := grpcmetadata.NewOutgoingContext(
		ctx,
		grpcmetadata.Pairs("authorization", "Bearer "+tokenResp.Token),
	)

	m := &gen.Metadata{
		Id:          "the-book",
		Title:       "The Book",
		Description: "The one and only book",
		Author:      "Mr. TJ",
		Isbn:        "123456789101",
	}

	// subtests (sequential flow)
	t.Run("Metadata Lifecycle", func(t *testing.T) {
		if _, err := metadataClient.PutMetadata(authCtx, &gen.PutMetadataRequest{Metadata: m}); err != nil {
			t.Fatalf("put metadata: %v", err)
		}

		getMetadataResp, err := metadataClient.GetMetadata(ctx, &gen.GetMetadataRequest{BookId: m.Id})
		if err != nil {
			t.Fatalf("get metadata: %v", err)
		}

		if diff := cmp.Diff(getMetadataResp.Metadata, m, cmpopts.IgnoreUnexported(gen.Metadata{})); diff != "" {
			t.Fatalf("get metadata after put mismatch:\n%s", diff)
		}
	})

	t.Run("Initial Book Details Verification", func(t *testing.T) {
		wantBookDetails := &gen.BookDetails{Metadata: m}

		getBookDetailsResp, err := bookClient.GetBookDetails(ctx, &gen.GetBookDetailsRequest{BookId: m.Id})
		if err != nil {
			t.Fatalf("get book details: %v", err)
		}

		if diff := cmp.Diff(getBookDetailsResp.BookDetails, wantBookDetails, cmpopts.IgnoreUnexported(gen.BookDetails{}, gen.Metadata{})); diff != "" {
			t.Fatalf("get book details after put mismatch:\n%s", diff)
		}
	})

	t.Run("Rating Aggregation Flow", func(t *testing.T) {
		const userID = "user0"
		const recordTypeBook = "book"
		firstRating := int32(5)
		secondRating := int32(1)
		wantRating := float64((firstRating + secondRating) / 2)

		if _, err := ratingClient.PutRating(authCtx, &gen.PutRatingRequest{UserId: userID, RecordId: m.Id, RecordType: recordTypeBook, RatingValue: firstRating}); err != nil {
			t.Fatalf("put rating 1: %v", err)
		}

		getAggregatedRatingResp, err := ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: m.Id, RecordType: recordTypeBook})
		if err != nil {
			t.Fatalf("get aggregated rating 1: %v", err)
		}
		if got, want := getAggregatedRatingResp.RatingValue, float64(5); got != want {
			t.Fatalf("rating mismatch. got %v want %v", got, want)
		}

		if _, err := ratingClient.PutRating(authCtx, &gen.PutRatingRequest{UserId: userID, RecordId: m.Id, RecordType: recordTypeBook, RatingValue: secondRating}); err != nil {
			t.Fatalf("put rating 2: %v", err)
		}

		getAggregatedRatingResp, err = ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: m.Id, RecordType: recordTypeBook})
		if err != nil {
			t.Fatalf("get aggregated rating 2: %v", err)
		}
		if got, want := getAggregatedRatingResp.RatingValue, wantRating; got != want {
			t.Fatalf("rating mismatch. got %v want %v", got, want)
		}
	})

	t.Run("Updated Book Details Verification", func(t *testing.T) {
		wantBookDetails := &gen.BookDetails{
			Metadata: m,
			Rating:   3.0,
		}

		getBookDetailsResp, err := bookClient.GetBookDetails(ctx, &gen.GetBookDetailsRequest{BookId: m.Id})
		if err != nil {
			t.Fatalf("get book details: %v", err)
		}

		if diff := cmp.Diff(getBookDetailsResp.BookDetails, wantBookDetails, cmpopts.IgnoreUnexported(gen.BookDetails{}, gen.Metadata{})); diff != "" {
			t.Fatalf("get movie details after update mismatch:\n%s", diff)
		}
	})
}
