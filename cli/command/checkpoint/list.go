package checkpoint

import (
	"context"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/spf13/cobra"
)

type listOptions struct {
	checkpointDir string
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS] CONTAINER",
		Aliases: []string{"list"},
		Short:   "List checkpoints for a container",
		Args:    cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, args[0], opts)
		},
		ValidArgsFunction: completion.ContainerNames(dockerCli, false),
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.checkpointDir, "checkpoint-dir", "", "", "Use a custom checkpoint storage directory")

	return cmd
}

func runList(dockerCli command.Cli, container string, opts listOptions) error {
	client := dockerCli.Client()

	listOpts := types.CheckpointListOptions{
		CheckpointDir: opts.checkpointDir,
	}

	checkpoints, err := client.CheckpointList(context.Background(), container, listOpts)
	if err != nil {
		return err
	}

	cpCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewFormat(formatter.TableFormatKey),
	}
	return FormatWrite(cpCtx, checkpoints)
}
