package client // import "github.com/DevanshMathur19/docker-v23/client"

import "context"

// NetworkRemove removes an existent network from the docker host.
func (cli *Client) NetworkRemove(ctx context.Context, networkID string) error {
	resp, err := cli.delete(ctx, "/networks/"+networkID, nil, nil)
	defer ensureReaderClosed(resp)
	return err
}
