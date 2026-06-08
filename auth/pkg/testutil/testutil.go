package testutil

import (
	grpchandler "github.com/tj330/bookapp/auth/internal/handler/grpc"
	"github.com/tj330/bookapp/gen"
)

// NewTestAuthGRPCServer returns a new AuthServiceServer
// with the test jwt-secret-key
func NewTestAuthGRPCServer() gen.AuthServiceServer {
	return grpchandler.New(func() []byte {
		return []byte("test-secret")
	})
}
