package main

import (
	"context"
	"net"
	"testing"

	authtest "github.com/tj330/bookapp/auth/pkg/testutil"
	booktest "github.com/tj330/bookapp/book/pkg/testutil"
	"github.com/tj330/bookapp/gen"
	metadatatest "github.com/tj330/bookapp/metadata/pkg/testutil"
	"github.com/tj330/bookapp/pkg/discovery"
	ratingtest "github.com/tj330/bookapp/rating/pkg/testutil"
	"google.golang.org/grpc"
)

// Shared environment configurations accessible by integration_test.go
const (
	metadataServiceName    = "metadata"
	ratingServiceName      = "rating"
	bookServiceName        = "book"
	metadataServiceAddress = "localhost:8081"
	ratingServiceAddress   = "localhost:8082"
	bookServiceAddress     = "localhost:8083"
	authServiceAddress     = "localhost:8084"
)

func startMetadataService(t *testing.T, ctx context.Context, registry discovery.Registry) *grpc.Server {
	t.Log("Starting metadata service on " + metadataServiceAddress)
	h := metadatatest.NewTestMetadataGRPCServer()

	l, err := net.Listen("tcp", metadataServiceAddress)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterMetadataServiceServer(srv, h)

	t.Cleanup(func() {
		srv.GracefulStop()
	})

	go func() {
		if err := srv.Serve(l); err != nil && err != grpc.ErrServerStopped {
			t.Logf("metadata gRPC server error: %v", err)
		}
	}()

	id := discovery.GenerateInstanceID(metadataServiceName)
	if err := registry.Register(ctx, id, metadataServiceName, metadataServiceAddress); err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	return srv
}

func startRatingService(t *testing.T, ctx context.Context, registry discovery.Registry) *grpc.Server {
	t.Log("Starting rating service on " + ratingServiceAddress)

	h := ratingtest.NewTestRatingGRPCServer()
	l, err := net.Listen("tcp", ratingServiceAddress)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterRatingServiceServer(srv, h)

	t.Cleanup(func() {
		srv.GracefulStop()
	})

	go func() {
		if err := srv.Serve(l); err != nil && err != grpc.ErrServerStopped {
			t.Logf("rating gRPC server error: %v", err)
		}
	}()

	id := discovery.GenerateInstanceID(ratingServiceName)
	if err := registry.Register(ctx, id, ratingServiceName, ratingServiceAddress); err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	return srv
}

func startBookService(t *testing.T, ctx context.Context, registry discovery.Registry) *grpc.Server {
	t.Log("Starting book service on " + bookServiceAddress)
	h := booktest.NewTestBookGRPCServer(registry)

	l, err := net.Listen("tcp", bookServiceAddress)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterBookServiceServer(srv, h)

	t.Cleanup(func() {
		srv.GracefulStop()
	})

	go func() {
		if err := srv.Serve(l); err != nil && err != grpc.ErrServerStopped {
			t.Logf("book gRPC server error: %v", err)
		}
	}()

	id := discovery.GenerateInstanceID(bookServiceName)
	if err := registry.Register(ctx, id, bookServiceName, bookServiceAddress); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	return srv
}

func startAuthService(t *testing.T, ctx context.Context, registry discovery.Registry) *grpc.Server {
	t.Log("Starting auth service on " + authServiceAddress)
	h := authtest.NewTestAuthGRPCServer()

	l, err := net.Listen("tcp", authServiceAddress)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterAuthServiceServer(srv, h)

	t.Cleanup(func() {
		srv.GracefulStop()
	})

	go func() {
		if err := srv.Serve(l); err != nil && err != grpc.ErrServerStopped {
			t.Logf("auth gRPC server error: %v", err)
		}
	}()

	id := discovery.GenerateInstanceID("auth")
	if err := registry.Register(ctx, id, "auth", authServiceAddress); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	return srv
}
