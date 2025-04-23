package client // import "github.com/harness-community/docker-v23/client"

import (
	"context"
	"net/url"

	"github.com/harness-community/docker-v23/api/types"
)

// PluginRemove removes a plugin
func (cli *Client) PluginRemove(ctx context.Context, name string, options types.PluginRemoveOptions) error {
	query := url.Values{}
	if options.Force {
		query.Set("force", "1")
	}

	resp, err := cli.delete(ctx, "/plugins/"+name, query, nil)
	defer ensureReaderClosed(resp)
	return err
}
