package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tj330/bookapp/book/internal/controller/book"
	metadataGateway "github.com/tj330/bookapp/book/internal/gateway/metadata/http"
	ratingGateway "github.com/tj330/bookapp/book/internal/gateway/rating/http"
	httphandler "github.com/tj330/bookapp/book/internal/handler/http"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/pkg/discovery/consul"
)

const serviceName = "book"

func main() {
	var port int
	flag.IntVar(&port, "port", 8083, "API hander port")
	flag.Parse()
	log.Printf("Starting the book service on port %d", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
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

	log.Println("Starting the book service")
	metadataGateway := metadataGateway.New(registry)
	ratingGateway := ratingGateway.New(registry)
	ctrl := book.New(ratingGateway, metadataGateway)
	h := httphandler.New(ctrl)
	http.Handle("/book", http.HandlerFunc(h.GetBookDetails))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
