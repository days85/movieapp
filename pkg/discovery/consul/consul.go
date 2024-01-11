package consul

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"movieexample.com/pkg/discovery"
)

type Registry struct {
	client *consul.Client
}

func NewRegistry(addr string) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Registry{client: client}, nil
}

func (r *Registry) Register(_ context.Context, instanceID string, serviceName string, hostPort string) error {
	addr := strings.Split(hostPort, ":")
	if len(addr) != 2 {
		return errors.New("hostPort must be in a form of <host>:<port>, example: localhost:8081")
	}

	port, err := strconv.Atoi(addr[1])
	if err != nil {
		return err
	}

	return r.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Address: addr[0],
		ID:      instanceID,
		Name:    serviceName,
		Port:    port,
		Check: &consul.AgentServiceCheck{
			CheckID: instanceID,
			TTL:     "5s",
		},
	})
}

func (r *Registry) Deregister(ctx context.Context, instanceID string, _ string) error {
	return r.client.Agent().ServiceDeregister(instanceID)
}

func (r *Registry) ServiceAddresses(_ context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, discovery.ErrNotFound
	}
	var res []string
	for _, e := range entries {
		res = append(res, fmt.Sprintf("%s:%d", e.Service.Address, e.Service.Port))
	}
	return res, nil
}

func (r *Registry) ReportHealthyState(instanceID string, _ string) error {
	return r.client.Agent().PassTTL(instanceID, "")
}
