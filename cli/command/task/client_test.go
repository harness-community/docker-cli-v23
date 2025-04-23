package task

import (
	"context"

	"github.com/DevanshMathur19/docker-v23/api/types"
	"github.com/DevanshMathur19/docker-v23/api/types/swarm"
	"github.com/DevanshMathur19/docker-v23/client"
)

type fakeClient struct {
	client.APIClient
	nodeInspectWithRaw    func(ref string) (swarm.Node, []byte, error)
	serviceInspectWithRaw func(ref string, options types.ServiceInspectOptions) (swarm.Service, []byte, error)
}

func (cli *fakeClient) NodeInspectWithRaw(ctx context.Context, ref string) (swarm.Node, []byte, error) {
	if cli.nodeInspectWithRaw != nil {
		return cli.nodeInspectWithRaw(ref)
	}
	return swarm.Node{}, nil, nil
}

func (cli *fakeClient) ServiceInspectWithRaw(ctx context.Context, ref string, options types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	if cli.serviceInspectWithRaw != nil {
		return cli.serviceInspectWithRaw(ref, options)
	}
	return swarm.Service{}, nil, nil
}
