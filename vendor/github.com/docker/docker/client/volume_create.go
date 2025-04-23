package client // import "github.com/DevanshMathur19/docker-v23/client"

import (
	"context"
	"encoding/json"

	"github.com/DevanshMathur19/docker-v23/api/types/volume"
)

// VolumeCreate creates a volume in the docker host.
func (cli *Client) VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error) {
	var vol volume.Volume
	resp, err := cli.post(ctx, "/volumes/create", nil, options, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		return vol, err
	}
	err = json.NewDecoder(resp.body).Decode(&vol)
	return vol, err
}
