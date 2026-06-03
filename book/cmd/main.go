package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/m3db/prometheus_client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tj330/bookapp/book/internal/controller/book"
	metadataGateway "github.com/tj330/bookapp/book/internal/gateway/metadata/grpc"
	ratingGateway "github.com/tj330/bookapp/book/internal/gateway/rating/grpc"
	grpchandler "github.com/tj330/bookapp/book/internal/handler/grpc"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
	"github.com/tj330/bookapp/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "book"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, err := os.Open("default.yml")
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	port := cfg.API.Port
	logger.Info("Starting the book service", zap.Int("port", port))
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
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
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("book:%d", port)); err != nil {
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

	certBytes, err := os.ReadFile("server.crt")
	if err != nil {
		logger.Fatal("Failed to read the server certificate", zap.Error(err))
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certBytes) {
		logger.Fatal("Failed to append server certificate to pool")
	}
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		logger.Fatal("Failed to load key pair", zap.Error(err))
	}
	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
	metadataGateway := metadataGateway.New(registry, creds)
	//metadataGateway := metadataGateway.New(registry)

	ratingGateway := ratingGateway.New(registry, creds)
	ctrl := book.New(ratingGateway, metadataGateway)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	const limit = 100
	const burst = 100
	l := newLimiter(100, 100)

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(
			ratelimit.UnaryServerInterceptor(l),
		),

		grpc.StatsHandler(
			otelgrpc.NewServerHandler(),
		),
	}

	srv := grpc.NewServer(opts...)

	reflection.Register(srv)

	gen.RegisterBookServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

type limiter struct {
	l *rate.Limiter
}

func newLimiter(limit int, burst int) *limiter {
	return &limiter{rate.NewLimiter(rate.Limit(limit), burst)}
}

func (l *limiter) Limit() bool {
	return !l.l.Allow()
}
