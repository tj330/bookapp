package authz

import (
	"context"
	"strings"

	"github.com/tj330/bookapp/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcmetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const usernameKey contextKey = "username"

func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}

func UnaryInterceptor(authClient gen.AuthServiceClient, protectedMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if !protectedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := grpcmetadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token, ok := strings.CutPrefix(values[0], "Bearer ")
		if !ok || token == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		resp, err := authClient.ValidateToken(ctx, &gen.ValidateTokenRequest{Token: token})
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, usernameKey, resp.Username)
		return handler(ctx, req)
	}
}
