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

type SecretProvider func() []byte

type Handler struct {
	secretProvider SecretProvider
	gen.UnimplementedAuthServiceServer
}

func New(secretProvider SecretProvider) *Handler {
	return &Handler{secretProvider: secretProvider}
}

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

func validCredentials(username string, password string) bool {
	if username == "" || password == "" {
		return false
	}
	// We intentionally skip verification of username and password
	// to simplify chapter code.
	return true
}

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
