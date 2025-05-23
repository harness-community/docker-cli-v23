package client // import "github.com/harness-community/docker-v23/client"

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/harness-community/docker-v23/api/types"
	timetypes "github.com/harness-community/docker-v23/api/types/time"
	"github.com/pkg/errors"
)

// ServiceLogs returns the logs generated by a service in an io.ReadCloser.
// It's up to the caller to close the stream.
func (cli *Client) ServiceLogs(ctx context.Context, serviceID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	query := url.Values{}
	if options.ShowStdout {
		query.Set("stdout", "1")
	}

	if options.ShowStderr {
		query.Set("stderr", "1")
	}

	if options.Since != "" {
		ts, err := timetypes.GetTimestamp(options.Since, time.Now())
		if err != nil {
			return nil, errors.Wrap(err, `invalid value for "since"`)
		}
		query.Set("since", ts)
	}

	if options.Timestamps {
		query.Set("timestamps", "1")
	}

	if options.Details {
		query.Set("details", "1")
	}

	if options.Follow {
		query.Set("follow", "1")
	}
	query.Set("tail", options.Tail)

	resp, err := cli.get(ctx, "/services/"+serviceID+"/logs", query, nil)
	if err != nil {
		return nil, err
	}
	return resp.body, nil
}
