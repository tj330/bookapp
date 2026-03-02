package memory

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/tj330/bookapp/pkg/discovery"
)

type Registry struct {
	sync.RWMutex
	serviceAddrs map[string]map[string]*serviceInstance
}

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

func New() *Registry {
	return &Registry{serviceAddrs: map[string]map[string]*serviceInstance{}}
}

func (r *Registry) Register(ctx context.Context, instanceId string, serviceName string, hostPort string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		r.serviceAddrs[serviceName] = map[string]*serviceInstance{}
	}
	r.serviceAddrs[serviceName][instanceId] = &serviceInstance{hostPort: hostPort, lastActive: time.Now()}
	return nil
}

func (r *Registry) Deregister(ctx context.Context, instanceId string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		return nil
	}
	delete(r.serviceAddrs[serviceName], instanceId)
	return nil
}

func (r *Registry) ReportHealthyState(instanceID string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		return errors.New("service is not registered yet")
	}
	if _, ok := r.serviceAddrs[serviceName][instanceID]; !ok {
		return errors.New("instance " + instanceID + " of service " + serviceName + " is not registered yet")
	}
	r.serviceAddrs[serviceName][instanceID].lastActive = time.Now()
	return nil
}

func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	r.Lock()
	defer r.Unlock()
	if len(r.serviceAddrs[serviceName]) == 0 {
		return nil, discovery.ErrNotFound
	}
	var res []string
	for instanceID, i := range r.serviceAddrs[serviceName] {
		if i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			log.Println("instance " + instanceID + " of service " + serviceName + " is not active, skipping")
			continue
		}
		res = append(res, i.hostPort)
	}
	return res, nil
}
