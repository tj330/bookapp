# BookApp

A Go-based microservice application for managing book metadata and ratings using Protocol Buffers and gRPC.

## Overview

BookApp is a distributed system designed to manage book information across multiple services. It provides three core gRPC services for handling metadata, ratings, and book details.

## Architecture
```
                    ┌─────────────────────────────┐
                    │  Consul Service Registry    │
                    └─────────────────────────────┘
                             ▲     ▲     ▲
                             │     │     │
                    ┌────────┘     │     └────────┐
                    │              │              │
                    ▼              ▼              ▼
            ┌───────────────┐ ┌──────────────┐ ┌──────────────┐
            │MetadataService│ │RatingService │ │ BookService  │
            │ (Port 8081)   │ │ (Port 8082)  │ │ (Port 8083)  │
            └──────┬────────┘ └──────┬───────┘ └──────────────┘
                   │                 │
                   │                 │
              PostgreSQL        ┌────▼────┐
              (Metadata)        │  Kafka  │
                                │ Consumer│
                                └────┬────┘
                                     │
                              ┌──────▼──────┐
                              │  Kafka      │
                              │  Topic:     │
                              │  'ratings'  │
                              └──────┬──────┘
                                     │
                                     │
                              ┌──────▼──────────┐
                              │ Rating Producer │
                              │ (cmd/           │
                              │  ratingproducer)│
                              └─────────────────┘
                                     │
                              ┌──────▼──────┐
                              │ PostgreSQL  │
                              │ (Ratings)   │
                              └─────────────┘
```
- **Language:** Go
- **Protocol:** Protocol Buffers (proto3)
- **Communication:** gRPC
- **Containerization:** Docker

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

**Kafka Integration**:

The service consumes rating events from the `ratings` topic (Kafka consumer group: `rating`). Each event is a JSON-encoded `RatingEvent`:

```json
{
  "userId": "105",
  "recordId": "1",
  "recordType": "movie",
  "value": 5,
  "providerId": "test-provider",
  "eventType": "put"
}
```

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

The services are designed to be consumed by gRPC clients and HTTP clients (for BookService). Example usage patterns:

### Get Book Details (Recommended Entry Point)

**gRPC:**
1. Create a BookService gRPC client
2. Call `GetBookDetails` with a book ID
3. Receive book metadata and current aggregated rating
4. Error handling: Returns `NotFound` if book doesn't exist

**HTTP:**
1. Make a GET request to the BookService HTTP endpoint
2. Provide the book ID as a query parameter: `?id=<book_id>`
3. Receive JSON response with book details

### Submit a Rating (Kafka Event Processing)

**Flow:**
1. Create a RatingService gRPC client
2. Call `PutRating` with:
   - `user_id`: Unique user identifier
   - `record_id`: The book ID
   - `record_type`: Type of record (e.g., "book")
   - `rating_value`: Integer rating (typically 1-5)
3. Rating is stored in PostgreSQL
4. **Kafka Integration:** The RatingService also consumes rating events from a Kafka topic (`ratings`), processes them asynchronously, and updates the aggregated ratings in the database

### Manage Metadata

1. Create a MetadataService gRPC client
2. **GetMetadata:** Query metadata by book ID
   - Parameters: `book_id`
   - Returns: Metadata object with title, author, description, ISBN
3. **PutMetadata:** Create or update book metadata
   - Parameters: Complete Metadata object with id, title, description, author, isbn
   - Returns: Empty response on success

### Integration Example

The typical workflow is:
1. Submit metadata via `MetadataService.PutMetadata`
2. Submit ratings via `RatingService.PutRating` or through Kafka events
3. Retrieve complete book details via `BookService.GetBookDetails` (which aggregates both services)
4. The aggregated rating is automatically calculated from all user ratings

### Service Discovery

All services register with Consul for service discovery:
- MetadataService: Port 8081
- RatingService: Port 8082
- BookService: Port 8083

Services automatically report health status to Consul every second.
