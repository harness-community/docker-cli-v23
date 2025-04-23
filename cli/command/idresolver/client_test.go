package idresolver

import (
	"context"

	"github.com/DevanshMathur19/docker-v23/api/types"
	"github.com/DevanshMathur19/docker-v23/api/types/swarm"
	"github.com/DevanshMathur19/docker-v23/client"
)

type fakeClient struct {
	client.Client
	nodeInspectFunc    func(string) (swarm.Node, []byte, error)
	serviceInspectFunc func(string) (swarm.Service, []byte, error)
}

func (cli *fakeClient) NodeInspectWithRaw(ctx context.Context, nodeID string) (swarm.Node, []byte, error) {
	if cli.nodeInspectFunc != nil {
		return cli.nodeInspectFunc(nodeID)
	}
	return swarm.Node{}, []byte{}, nil
}

func (cli *fakeClient) ServiceInspectWithRaw(ctx context.Context, serviceID string, options types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	if cli.serviceInspectFunc != nil {
		return cli.serviceInspectFunc(serviceID)
	}
	return swarm.Service{}, []byte{}, nil
}
