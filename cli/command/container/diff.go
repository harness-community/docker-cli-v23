package container

import (
	"context"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type diffOptions struct {
	container string
}

// NewDiffCommand creates a new cobra.Command for `docker diff`
func NewDiffCommand(dockerCli command.Cli) *cobra.Command {
	var opts diffOptions

	return &cobra.Command{
		Use:   "diff CONTAINER",
		Short: "Inspect changes to files or directories on a container's filesystem",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.container = args[0]
			return runDiff(dockerCli, &opts)
		},
		Annotations: map[string]string{
			"aliases": "docker container diff, docker diff",
		},
		ValidArgsFunction: completion.ContainerNames(dockerCli, false),
	}
}

func runDiff(dockerCli command.Cli, opts *diffOptions) error {
	if opts.container == "" {
		return errors.New("Container name cannot be empty")
	}
	ctx := context.Background()

	changes, err := dockerCli.Client().ContainerDiff(ctx, opts.container)
	if err != nil {
		return err
	}
	diffCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewDiffFormat("{{.Type}} {{.Path}}"),
	}
	return DiffFormatWrite(diffCtx, changes)
}
