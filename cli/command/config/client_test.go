package config

import (
	"context"

	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/harness-community/docker-v23/client"
)

type fakeClient struct {
	client.Client
	configCreateFunc  func(swarm.ConfigSpec) (types.ConfigCreateResponse, error)
	configInspectFunc func(string) (swarm.Config, []byte, error)
	configListFunc    func(types.ConfigListOptions) ([]swarm.Config, error)
	configRemoveFunc  func(string) error
}

func (c *fakeClient) ConfigCreate(ctx context.Context, spec swarm.ConfigSpec) (types.ConfigCreateResponse, error) {
	if c.configCreateFunc != nil {
		return c.configCreateFunc(spec)
	}
	return types.ConfigCreateResponse{}, nil
}

func (c *fakeClient) ConfigInspectWithRaw(ctx context.Context, id string) (swarm.Config, []byte, error) {
	if c.configInspectFunc != nil {
		return c.configInspectFunc(id)
	}
	return swarm.Config{}, nil, nil
}

func (c *fakeClient) ConfigList(ctx context.Context, options types.ConfigListOptions) ([]swarm.Config, error) {
	if c.configListFunc != nil {
		return c.configListFunc(options)
	}
	return []swarm.Config{}, nil
}

func (c *fakeClient) ConfigRemove(ctx context.Context, name string) error {
	if c.configRemoveFunc != nil {
		return c.configRemoveFunc(name)
	}
	return nil
}
