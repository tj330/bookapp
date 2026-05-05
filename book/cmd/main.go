package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tj330/bookapp/book/internal/controller/book"
	metadataGateway "github.com/tj330/bookapp/book/internal/gateway/metadata/grpc"
	ratingGateway "github.com/tj330/bookapp/book/internal/gateway/rating/grpc"
	grpchandler "github.com/tj330/bookapp/book/internal/handler/grpc"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v3"
)

const serviceName = "book"

func main() {
	f, err := os.Open("default.yml")
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	port := cfg.API.Port
	log.Printf("Starting the book service on port %d", port)
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("book:%d", port)); err != nil {
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

	certBytes, err := os.ReadFile("server.crt")
	if err != nil {
		log.Fatalf("failed to read the server certificate: %v", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certBytes) {
		log.Fatalf("failed to append server certificate to pool")
	}
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}
	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
	metadataGateway := metadataGateway.New(registry, creds)
	ratingGateway := ratingGateway.New(registry)
	ctrl := book.New(ratingGateway, metadataGateway)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	srv := grpc.NewServer(
		grpc.Creds(creds),
	)
	gen.RegisterBookServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
