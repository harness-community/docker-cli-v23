package client // import "github.com/harness-community/docker-v23/client"

import (
	"context"
	"net/url"

	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/harness-community/docker-v23/api/types/volume"
)

// VolumeUpdate updates a volume. This only works for Cluster Volumes, and
// only some fields can be updated.
func (cli *Client) VolumeUpdate(ctx context.Context, volumeID string, version swarm.Version, options volume.UpdateOptions) error {
	if err := cli.NewVersionError("1.42", "volume update"); err != nil {
		return err
	}

	query := url.Values{}
	query.Set("version", version.String())

	resp, err := cli.put(ctx, "/volumes/"+volumeID, query, options, nil)
	ensureReaderClosed(resp)
	return err
}
