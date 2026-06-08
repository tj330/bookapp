package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Reading and extracting configured values from default.yml.
	f, err := os.Open("default.yml")
	if err != nil {
		panic(err)
	}
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	port := cfg.API.Port
	logger.Info("Starting the rating service", zap.Int("port", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger.Info("Jaeger configured",
		zap.String("endpoint", cfg.Jaeger.URL),
	)

	// Setting up tracing for the service using jaeger.
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

	// Counter to mark the no of
	// instance of the service.
	serviceStartedCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_started_total",
			Help: "Number of times the service started",
		},
		[]string{"service"},
	)

	prometheus.MustRegister(serviceStartedCounter)

	// HTTP endpoint for exposing metrics
	http.Handle("/metrics", promhttp.Handler())

	// Go-routine for exposing metrics
	// on the configured port.
	go func() {
		if err := http.ListenAndServe(
			fmt.Sprintf(":%d", cfg.Prometheus.MetricsPort),
			nil,
		); err != nil {
			logger.Fatal(
				"Failed to start metrics handler",
				zap.Error(err),
			)
		}
	}()

	serviceStartedCounter.
		WithLabelValues(serviceName).
		Inc()

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}

	// Register the service with Consul so that other services can discover it.
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("rating:%d", port)); err != nil {
		panic(err)
	}

	// Go-routine to report healthy state to Consul.
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	// asynchronously consuming rating using kafka from topic `ratings`.
	ingester, err := kafka.NewIngester("localhost", "rating", "ratings")
	if err != nil {
		logger.Fatal("Failed to initialize ingester", zap.Error(err))
	}

	repo, err := psql.New()
	if err != nil {
		panic(err)
	}
	// ctrl := rating.New(repo)
	ctrl := rating.New(repo, ingester)

	// Go-routine for async consumption.
	go func() {
		if err := ctrl.StartIngestion(ctx); err != nil {
			logger.Fatal("Failed to start ingestion", zap.Error(err))
		}
	}()

	h := grpcphandler.New(ctrl)

	// Loading the server certificate and private key.
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		logger.Fatal("Failed to load key pair", zap.Error(err))
	}

	// Building the transport credentials based on tls and listening on the
	// configured port.
	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})

	// Connecting with the auth service.
	authConn, err := grpcutil.ServiceConnection(ctx, "auth", registry, creds)
	if err != nil {
		logger.Fatal("Failed to connect to auth service", zap.Error(err))
	}
	defer authConn.Close()

	authClient := gen.NewAuthServiceClient(authConn)
	// Defining methods that require authentication
	protectedMethods := map[string]bool{
		"/RatingService/PutRating": true,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	// gRPC server option including the auth interceptor.
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(authz.UnaryInterceptor(authClient, protectedMethods)),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}

	// Creating new gRPC server with
	// corresponding tls credentials.
	srv := grpc.NewServer(opts...)
	reflection.Register(srv)

	// Setting up signal channel for handling operating system
	// signals to execute graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Registering rating service with the server and
	// custom handler.
	gen.RegisterRatingServiceServer(srv, h)
	go func() {
		s := <-sigChan
		cancel()
		logger.Info("Received signal, attempting graceful shutdown", zap.Stringer("signal", s))
		srv.GracefulStop()
		logger.Info("Gracefully stopped the gRPC server")
	}()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
