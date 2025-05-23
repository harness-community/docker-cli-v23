package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	force bool
}

func newRemoveCommand(dockerCli command.Cli) *cobra.Command {
	opts := removeOptions{}

	cmd := &cobra.Command{
		Use:     "rm [OPTIONS] NODE [NODE...]",
		Aliases: []string{"remove"},
		Short:   "Remove one or more nodes from the swarm",
		Args:    cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(dockerCli, args, opts)
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&opts.force, "force", "f", false, "Force remove a node from the swarm")
	return cmd
}

func runRemove(dockerCli command.Cli, args []string, opts removeOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	var errs []string

	for _, nodeID := range args {
		err := client.NodeRemove(ctx, nodeID, types.NodeRemoveOptions{Force: opts.force})
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		fmt.Fprintf(dockerCli.Out(), "%s\n", nodeID)
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, "\n"))
	}

	return nil
}
