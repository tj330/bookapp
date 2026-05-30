package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/internal/authz"
	"github.com/tj330/bookapp/internal/grpcutil"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
	"github.com/tj330/bookapp/pkg/tracing"
	"github.com/tj330/bookapp/rating/internal/controller/rating"
	grpcphandler "github.com/tj330/bookapp/rating/internal/handler/grpc"
	"github.com/tj330/bookapp/rating/internal/ingester/kafka"
	"github.com/tj330/bookapp/rating/internal/repository/psql"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "rating"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	f, err := os.Open("default.yml")
	if err != nil {
		panic(err)
	}
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	port := cfg.API.Port
	log.Printf("Starting the rating service on port %d", port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Printf("Jaeger URL = %q\n", cfg.Jaeger.URL)
	tp, err := tracing.NewJaegerProvider(cfg.Jaeger.URL, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize Jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Fatal("Failed to shutdown Jaeger provider", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("rating:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	ingester, err := kafka.NewIngester("localhost", "rating", "ratings")
	if err != nil {
		log.Fatalf("failed to initialize ingester: %v", err)
	}

	repo, err := psql.New()
	if err != nil {
		panic(err)
	}
	// ctrl := rating.New(repo)
	ctrl := rating.New(repo, ingester)

	go func() {
		if err := ctrl.StartIngestion(ctx); err != nil {
			log.Fatalf("failed to start ingestion: %v", err)
		}
	}()

	h := grpcphandler.New(ctrl)

	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}

	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})

	authConn, err := grpcutil.ServiceConnection(ctx, "auth", registry, creds)
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer authConn.Close()

	authClient := gen.NewAuthServiceClient(authConn)

	protectedMethods := map[string]bool{
		"/RatingService/PutRating": true,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(authz.UnaryInterceptor(authClient, protectedMethods)),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}

	srv := grpc.NewServer(opts...)
	reflection.Register(srv)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	gen.RegisterRatingServiceServer(srv, h)
	go func() {
		s := <-sigChan
		cancel()
		log.Printf("Received signal %v, attempting graceful shutdown", s)
		srv.GracefulStop()
		log.Printf("Gracefully stopped the gRPC server")
	}()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
