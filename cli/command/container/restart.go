package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-v23/api/types/container"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type restartOptions struct {
	signal         string
	timeout        int
	timeoutChanged bool

	containers []string
}

// NewRestartCommand creates a new cobra.Command for `docker restart`
func NewRestartCommand(dockerCli command.Cli) *cobra.Command {
	var opts restartOptions

	cmd := &cobra.Command{
		Use:   "restart [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Restart one or more containers",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.containers = args
			opts.timeoutChanged = cmd.Flags().Changed("time")
			return runRestart(dockerCli, &opts)
		},
		Annotations: map[string]string{
			"aliases": "docker container restart, docker restart",
		},
		ValidArgsFunction: completion.ContainerNames(dockerCli, true),
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.signal, "signal", "s", "", "Signal to send to the container")
	flags.IntVarP(&opts.timeout, "time", "t", 0, "Seconds to wait before killing the container")
	return cmd
}

func runRestart(dockerCli command.Cli, opts *restartOptions) error {
	ctx := context.Background()
	var errs []string
	var timeout *int
	if opts.timeoutChanged {
		timeout = &opts.timeout
	}
	for _, name := range opts.containers {
		err := dockerCli.Client().ContainerRestart(ctx, name, container.StopOptions{
			Signal:  opts.signal,
			Timeout: timeout,
		})
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		_, _ = fmt.Fprintln(dockerCli.Out(), name)
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
