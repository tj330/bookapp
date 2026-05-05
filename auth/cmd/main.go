package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	grpchandler "github.com/tj330/bookapp/auth/internal/handler/grpc"
	"github.com/tj330/bookapp/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := 8084
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
