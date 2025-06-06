package client // import "github.com/harness-community/docker-v23/client"

import "context"

// ServiceRemove kills and removes a service.
func (cli *Client) ServiceRemove(ctx context.Context, serviceID string) error {
	resp, err := cli.delete(ctx, "/services/"+serviceID, nil, nil)
	defer ensureReaderClosed(resp)
	return err
}
