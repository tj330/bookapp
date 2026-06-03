package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/m3db/prometheus_client_golang/prometheus"
	"github.com/m3db/prometheus_client_golang/prometheus/promhttp"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/internal/authz"
	"github.com/tj330/bookapp/internal/grpcutil"
	"github.com/tj330/bookapp/metadata/internal/controller/metadata"
	grpchandler "github.com/tj330/bookapp/metadata/internal/handler/grpc"
	"github.com/tj330/bookapp/metadata/internal/repository/psql"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
	"github.com/tj330/bookapp/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "metadata"

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

	logger.Info("Starting the book metadata service", zap.Int("port", port))
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Addr)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("Jaeger configured",
		zap.String("endpoint", cfg.Jaeger.URL),
	)

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

	serviceStartedCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_started_total",
			Help: "Number of times the service started",
		},
		[]string{"service"},
	)

	prometheus.MustRegister(serviceStartedCounter)

	http.Handle("/metrics", promhttp.Handler())

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

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("metadata:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	repo, err := psql.New()
	if err != nil {
		panic(err)
	}
	ctrl := metadata.New(repo)
	h := grpchandler.New(ctrl)
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		logger.Fatal("Failed to load key pair", zap.Error(err))
	}

	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	authConn, err := grpcutil.ServiceConnection(ctx, "auth", registry, creds)
	if err != nil {
		logger.Fatal("Failed to connect to auth service", zap.Error(err))
	}
	defer authConn.Close()

	authClient := gen.NewAuthServiceClient(authConn)

	protectedMethods := map[string]bool{
		"/MetadataService/PutMetadata": true,
	}

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(authz.UnaryInterceptor(authClient, protectedMethods)),
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(),
		),
	}
	srv := grpc.NewServer(opts...)
	reflection.Register(srv)

	gen.RegisterMetadataServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
