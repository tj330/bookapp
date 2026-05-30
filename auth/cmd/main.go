package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	grpchandler "github.com/tj330/bookapp/auth/internal/handler/grpc"
	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := 8084
	ctx := context.Background()

	registry, err := consul.NewRegistry("http://localhost:8500")
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID("auth")
	if err := registry.Register(ctx, instanceID, "auth", fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	defer registry.Deregister(ctx, instanceID, "auth")

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, "auth"); err != nil {
				log.Println("Failed to report auth healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()

	log.Printf("Starting the auth service on port: %d", port)
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}

	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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
