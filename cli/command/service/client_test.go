package service

import (
	"context"

	. "github.com/harness-community/docker-cli-v23/internal/test/builders" // Import builders to get the builder function as package function
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/harness-community/docker-v23/client"
)

type fakeClient struct {
	client.Client
	serviceInspectWithRawFunc func(ctx context.Context, serviceID string, options types.ServiceInspectOptions) (swarm.Service, []byte, error)
	serviceUpdateFunc         func(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error)
	serviceListFunc           func(context.Context, types.ServiceListOptions) ([]swarm.Service, error)
	taskListFunc              func(context.Context, types.TaskListOptions) ([]swarm.Task, error)
	infoFunc                  func(ctx context.Context) (types.Info, error)
	networkInspectFunc        func(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
	nodeListFunc              func(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error)
}

func (f *fakeClient) NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error) {
	if f.nodeListFunc != nil {
		return f.nodeListFunc(ctx, options)
	}
	return nil, nil
}

func (f *fakeClient) TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error) {
	if f.taskListFunc != nil {
		return f.taskListFunc(ctx, options)
	}
	return nil, nil
}

func (f *fakeClient) ServiceInspectWithRaw(ctx context.Context, serviceID string, options types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	if f.serviceInspectWithRawFunc != nil {
		return f.serviceInspectWithRawFunc(ctx, serviceID, options)
	}

	return *Service(ServiceID(serviceID)), []byte{}, nil
}

func (f *fakeClient) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	if f.serviceListFunc != nil {
		return f.serviceListFunc(ctx, options)
	}

	return nil, nil
}

func (f *fakeClient) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error) {
	if f.serviceUpdateFunc != nil {
		return f.serviceUpdateFunc(ctx, serviceID, version, service, options)
	}

	return types.ServiceUpdateResponse{}, nil
}

func (f *fakeClient) Info(ctx context.Context) (types.Info, error) {
	if f.infoFunc == nil {
		return types.Info{}, nil
	}
	return f.infoFunc(ctx)
}

func (f *fakeClient) NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error) {
	if f.networkInspectFunc != nil {
		return f.networkInspectFunc(ctx, networkID, options)
	}
	return types.NetworkResource{}, nil
}

func newService(id string, name string) swarm.Service {
	return *Service(ServiceID(id), ServiceName(name))
}
