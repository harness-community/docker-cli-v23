package client // import "github.com/harness-community/docker-v23/client"

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/harness-community/docker-v23/api/types"
)

// PluginCreate creates a plugin
func (cli *Client) PluginCreate(ctx context.Context, createContext io.Reader, createOptions types.PluginCreateOptions) error {
	headers := http.Header(make(map[string][]string))
	headers.Set("Content-Type", "application/x-tar")

	query := url.Values{}
	query.Set("name", createOptions.RepoName)

	resp, err := cli.postRaw(ctx, "/plugins/create", query, createContext, headers)
	ensureReaderClosed(resp)
	return err
}
