package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tj330/bookapp/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Dependency injection pattern so that
// you can simply provide the secret
// from different places like, as environment
// variable, from a vault etc
type SecretProvider func() []byte

// Our custom handler including the secret provider
// and the generated service struct for forward compatability
// when we upgrade our api, missing our own custom implementation.
type Handler struct {
	secretProvider SecretProvider
	gen.UnimplementedAuthServiceServer
}

// New creates a Handler with the provided secrete provider.
func New(secretProvider SecretProvider) *Handler {
	return &Handler{secretProvider: secretProvider}
}

// GetToken authenticates the user and returns a signed JWT containing
// the username and issued-at timestamp.
func (h *Handler) GetToken(ctx context.Context, req *gen.GetTokenRequest) (*gen.GetTokenResponse, error) {
	username, password := req.Username, req.Password
	if !validCredentials(username, password) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString(h.secretProvider())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &gen.GetTokenResponse{Token: tokenString}, nil
}

// validCredentials returns true if the credentials are valid.
//
// Simple implementation for this project now, but can be tied with
// business logic
func validCredentials(username string, password string) bool {
	if username == "" || password == "" {
		return false
	}
	return true
}

// ValidateToken validates the JWT signature and extracts the username claim.
func (h *Handler) ValidateToken(ctx context.Context, req *gen.ValidateTokenRequest) (*gen.ValidateTokenResponse, error) {
	token, err := jwt.Parse(
		req.Token,
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return h.secretProvider(), nil
		})

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	var username string
	if v, ok := claims["username"]; ok {
		if u, ok := v.(string); ok {
			username = u
		}
	}

	return &gen.ValidateTokenResponse{Username: username}, nil
}
