package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
	"github.com/tj330/bookapp/rating/internal/controller/rating"
	grpcphandler "github.com/tj330/bookapp/rating/internal/handler/grpc"
	"github.com/tj330/bookapp/rating/internal/ingester/kafka"
	"github.com/tj330/bookapp/rating/internal/repository/psql"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

const serviceName = "rating"

func main() {
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
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("rating:%d", port)); err != nil {
		panic(err)
	}
	fmt.Print("Test")
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
	ctrl := rating.New(repo, ingester)
	go func() {
		if err := ctrl.StartIngestion(ctx); err != nil {
			log.Fatalf("failed to start ingestion: %v", err)
		}
	}()
	h := grpcphandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterRatingServiceServer(srv, h)

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
