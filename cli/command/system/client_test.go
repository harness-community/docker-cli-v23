package system

import (
	"context"

	"github.com/DevanshMathur19/docker-v23/api/types"
	"github.com/DevanshMathur19/docker-v23/client"
)

type fakeClient struct {
	client.Client

	version       string
	serverVersion func(ctx context.Context) (types.Version, error)
}

func (cli *fakeClient) ServerVersion(ctx context.Context) (types.Version, error) {
	return cli.serverVersion(ctx)
}

func (cli *fakeClient) ClientVersion() string {
	return cli.version
}
