package client // import "github.com/DevanshMathur19/docker-v23/client"

import (
	"context"

	"github.com/DevanshMathur19/docker-v23/api/types/swarm"
)

// SwarmJoin joins the swarm.
func (cli *Client) SwarmJoin(ctx context.Context, req swarm.JoinRequest) error {
	resp, err := cli.post(ctx, "/swarm/join", nil, req, nil)
	ensureReaderClosed(resp)
	return err
}
