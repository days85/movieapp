package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"movieexample.com/pkg/discovery"
)

type serviceName string
type instanceID string

type Registry struct {
	sync.RWMutex
	serviceAddrs map[serviceName]map[instanceID]*serviceInstance
}

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

func NewRegistry() *Registry {
	return &Registry{serviceAddrs: map[serviceName]map[instanceID]*serviceInstance{}}
}

func (r *Registry) Register(_ context.Context, id instanceID, name serviceName, hostPort string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[name]; !ok {
		r.serviceAddrs[name] = map[instanceID]*serviceInstance{}
	}
	r.serviceAddrs[name][id] = &serviceInstance{hostPort: hostPort, lastActive: time.Now()}
	return nil
}

func (r *Registry) Deregister(_ context.Context, id instanceID, name serviceName) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[name]; !ok {
		return nil
	}
	delete(r.serviceAddrs[name], id)
	return nil
}

func (r *Registry) ReportHealthyState(id instanceID, name serviceName) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[name]; !ok {
		return errors.New("service is not registered yet")
	}
	if _, ok := r.serviceAddrs[name][id]; !ok {
		return errors.New("service instance is not registered yet")
	}
	r.serviceAddrs[name][id].lastActive = time.Now()
	return nil
}

func (r *Registry) ServiceAddresses(_ context.Context, name serviceName) ([]string, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.serviceAddrs[name]) == 0 {
		return nil, discovery.ErrNotFound
	}

	var res []string

	for _, i := range r.serviceAddrs[name] {
		if i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			continue
		}
		res = append(res, i.hostPort)
	}
	return res, nil
}
