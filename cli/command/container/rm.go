package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/errdefs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type rmOptions struct {
	rmVolumes bool
	rmLink    bool
	force     bool

	containers []string
}

// NewRmCommand creates a new cobra.Command for `docker rm`
func NewRmCommand(dockerCli command.Cli) *cobra.Command {
	var opts rmOptions

	cmd := &cobra.Command{
		Use:     "rm [OPTIONS] CONTAINER [CONTAINER...]",
		Aliases: []string{"remove"},
		Short:   "Remove one or more containers",
		Args:    cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.containers = args
			return runRm(dockerCli, &opts)
		},
		Annotations: map[string]string{
			"aliases": "docker container rm, docker container remove, docker rm",
		},
		ValidArgsFunction: completion.ContainerNames(dockerCli, true),
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.rmVolumes, "volumes", "v", false, "Remove anonymous volumes associated with the container")
	flags.BoolVarP(&opts.rmLink, "link", "l", false, "Remove the specified link")
	flags.BoolVarP(&opts.force, "force", "f", false, "Force the removal of a running container (uses SIGKILL)")
	return cmd
}

func runRm(dockerCli command.Cli, opts *rmOptions) error {
	ctx := context.Background()

	var errs []string
	options := types.ContainerRemoveOptions{
		RemoveVolumes: opts.rmVolumes,
		RemoveLinks:   opts.rmLink,
		Force:         opts.force,
	}

	errChan := parallelOperation(ctx, opts.containers, func(ctx context.Context, container string) error {
		container = strings.Trim(container, "/")
		if container == "" {
			return errors.New("Container name cannot be empty")
		}
		return dockerCli.Client().ContainerRemove(ctx, container, options)
	})

	for _, name := range opts.containers {
		if err := <-errChan; err != nil {
			if opts.force && errdefs.IsNotFound(err) {
				fmt.Fprintln(dockerCli.Err(), err)
				continue
			}
			errs = append(errs, err.Error())
			continue
		}
		fmt.Fprintln(dockerCli.Out(), name)
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
