package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/m3db/prometheus_client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus"
	grpchandler "github.com/tj330/bookapp/auth/internal/handler/grpc"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
	"github.com/tj330/bookapp/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "auth"

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
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	defer registry.Deregister(ctx, instanceID, serviceName)

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()

	logger.Info("Starting the auth service", zap.Int("port", port))
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		logger.Fatal("Failed to load key pair", zap.Error(err))
	}

	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	h := grpchandler.New(func() []byte {
		return []byte("test-secret")
	})
	srv := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(srv)
	gen.RegisterAuthServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
