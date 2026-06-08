package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Registry is a service registry interface including the necessary methods.
type Registry interface {
	Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error
	Deregister(ctx context.Context, instanceID string, serviceName string) error
	ServiceAddresses(ctx context.Context, serviceID string) ([]string, error)
	ReportHealthyState(instanceID string, serviceName string) error
}

// Custom error if a service is not found.
var ErrNotFound = errors.New("service not found")

// Used to generate a random instance id using the serviceName.
func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
