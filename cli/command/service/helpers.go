package service

import (
	"context"
	"io"

	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/service/progress"
	"github.com/harness-community/docker-v23/pkg/jsonmessage"
)

// waitOnService waits for the service to converge. It outputs a progress bar,
// if appropriate based on the CLI flags.
func waitOnService(ctx context.Context, dockerCli command.Cli, serviceID string, quiet bool) error {
	errChan := make(chan error, 1)
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		errChan <- progress.ServiceProgress(ctx, dockerCli.Client(), serviceID, pipeWriter)
	}()

	if quiet {
		go io.Copy(io.Discard, pipeReader)
		return <-errChan
	}

	err := jsonmessage.DisplayJSONMessagesToStream(pipeReader, dockerCli.Out(), nil)
	if err == nil {
		err = <-errChan
	}
	return err
}
