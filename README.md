# BookApp

A Go-based microservice application for managing book metadata and ratings using Protocol Buffers and gRPC.

## Overview

BookApp is a distributed system designed to manage book information across multiple services. It provides three core gRPC services for handling metadata, ratings, and book details.

## Services

### MetadataService
Manages book metadata information including title, author, ISBN, and descriptions.

**RPCs:**
- `GetMetadata(GetMetadataRequest)` - Retrieve book metadata by book ID
- `PutMetadata(PutMetadataRequest)` - Create or update book metadata

### RatingService
Handles user ratings and aggregated rating calculations for books and other records.

**RPCs:**
- `GetAggregatedRating(GetAggregatedRatingRequest)` - Get aggregated rating for a record
- `PutRating(PutRatingRequest)` - Submit a user rating

### BookService
Provides comprehensive book details combining metadata and rating information.

**RPCs:**
- `GetBookDetails(GetBookDetailsRequest)` - Retrieve complete book details (metadata + rating)

## Data Models

### Metadata
- id: unique identifier
- title: book title
- description: book description
- author: author name
- isbn: ISBN number


### BookDetails
- rating: aggregated rating score
- metadata: associated metadata object


## Architecture

- **Language:** Go
- **Protocol:** Protocol Buffers (proto3)
- **Communication:** gRPC
- **Containerization:** Docker

## Getting Started

### Prerequisites
- Go 1.16+
- Protocol Buffers compiler (`protoc`)
- Docker (optional)

### Code Generation

Generate Go code from protobuf definitions:

```bash
protoc --go_out=. --go-grpc_out=. api/book.proto
```
This will generate code in the /gen directory as specified in the proto file.

### Building

```bash:
go build ./...
```

### Docker

Build the Docker image:
```bash:
docker build -t bookapp .
```


## Usage

The services are designed to be consumed by gRPC clients. Example usage patterns:

Get Book Details:
1. Create a BookService client
2. Call GetBookDetails with a book ID
3. Receive book metadata and aggregated rating

Submit a Rating:
1. Create a RatingService client
2. Call PutRating with user ID, record ID, record type, and rating value
3. Confirmation returned

Manage Metadata:
1. Create a MetadataService client
2. Use GetMetadata or PutMetadata as needed
