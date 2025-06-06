package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter/tabwriter"
	"github.com/spf13/cobra"
)

type topOptions struct {
	container string

	args []string
}

// NewTopCommand creates a new cobra.Command for `docker top`
func NewTopCommand(dockerCli command.Cli) *cobra.Command {
	var opts topOptions

	cmd := &cobra.Command{
		Use:   "top CONTAINER [ps OPTIONS]",
		Short: "Display the running processes of a container",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.container = args[0]
			opts.args = args[1:]
			return runTop(dockerCli, &opts)
		},
		Annotations: map[string]string{
			"aliases": "docker container top, docker top",
		},
		ValidArgsFunction: completion.ContainerNames(dockerCli, false),
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	return cmd
}

func runTop(dockerCli command.Cli, opts *topOptions) error {
	ctx := context.Background()

	procList, err := dockerCli.Client().ContainerTop(ctx, opts.container, opts.args)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(dockerCli.Out(), 20, 1, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(procList.Titles, "\t"))

	for _, proc := range procList.Processes {
		fmt.Fprintln(w, strings.Join(proc, "\t"))
	}
	w.Flush()
	return nil
}
