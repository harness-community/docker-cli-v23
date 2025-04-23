package client // import "github.com/DevanshMathur19/docker-v23/client"

import (
	"context"
	"net/url"

	"github.com/DevanshMathur19/docker-v23/api/types"
)

// NodeRemove removes a Node.
func (cli *Client) NodeRemove(ctx context.Context, nodeID string, options types.NodeRemoveOptions) error {
	query := url.Values{}
	if options.Force {
		query.Set("force", "1")
	}

	resp, err := cli.delete(ctx, "/nodes/"+nodeID, query, nil)
	defer ensureReaderClosed(resp)
	return err
}
