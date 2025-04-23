package client // import "github.com/harness-community/docker-v23/client"

import (
	"context"
	"net/url"
	"strconv"

	"github.com/harness-community/docker-v23/api/types"
)

// PluginEnable enables a plugin
func (cli *Client) PluginEnable(ctx context.Context, name string, options types.PluginEnableOptions) error {
	query := url.Values{}
	query.Set("timeout", strconv.Itoa(options.Timeout))

	resp, err := cli.post(ctx, "/plugins/"+name+"/enable", query, nil, nil)
	ensureReaderClosed(resp)
	return err
}
