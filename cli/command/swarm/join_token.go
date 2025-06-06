package swarm

import (
	"context"
	"fmt"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type joinTokenOptions struct {
	role   string
	rotate bool
	quiet  bool
}

func newJoinTokenCommand(dockerCli command.Cli) *cobra.Command {
	opts := joinTokenOptions{}

	cmd := &cobra.Command{
		Use:   "join-token [OPTIONS] (worker|manager)",
		Short: "Manage join tokens",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.role = args[0]
			return runJoinToken(dockerCli, opts)
		},
		Annotations: map[string]string{
			"version": "1.24",
			"swarm":   "manager",
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.rotate, flagRotate, false, "Rotate join token")
	flags.BoolVarP(&opts.quiet, flagQuiet, "q", false, "Only display token")

	return cmd
}

func runJoinToken(dockerCli command.Cli, opts joinTokenOptions) error {
	worker := opts.role == "worker"
	manager := opts.role == "manager"

	if !worker && !manager {
		return errors.New("unknown role " + opts.role)
	}

	client := dockerCli.Client()
	ctx := context.Background()

	if opts.rotate {
		flags := swarm.UpdateFlags{
			RotateWorkerToken:  worker,
			RotateManagerToken: manager,
		}

		sw, err := client.SwarmInspect(ctx)
		if err != nil {
			return err
		}

		if err := client.SwarmUpdate(ctx, sw.Version, sw.Spec, flags); err != nil {
			return err
		}

		if !opts.quiet {
			fmt.Fprintf(dockerCli.Out(), "Successfully rotated %s join token.\n\n", opts.role)
		}
	}

	// second SwarmInspect in this function,
	// this is necessary since SwarmUpdate after first changes the join tokens
	sw, err := client.SwarmInspect(ctx)
	if err != nil {
		return err
	}

	if opts.quiet && worker {
		fmt.Fprintln(dockerCli.Out(), sw.JoinTokens.Worker)
		return nil
	}

	if opts.quiet && manager {
		fmt.Fprintln(dockerCli.Out(), sw.JoinTokens.Manager)
		return nil
	}

	info, err := client.Info(ctx)
	if err != nil {
		return err
	}

	return printJoinCommand(ctx, dockerCli, info.Swarm.NodeID, worker, manager)
}

func printJoinCommand(ctx context.Context, dockerCli command.Cli, nodeID string, worker bool, manager bool) error {
	client := dockerCli.Client()

	node, _, err := client.NodeInspectWithRaw(ctx, nodeID)
	if err != nil {
		return err
	}

	sw, err := client.SwarmInspect(ctx)
	if err != nil {
		return err
	}

	if node.ManagerStatus != nil {
		if worker {
			fmt.Fprintf(dockerCli.Out(), "To add a worker to this swarm, run the following command:\n\n    docker swarm join --token %s %s\n\n", sw.JoinTokens.Worker, node.ManagerStatus.Addr)
		}
		if manager {
			fmt.Fprintf(dockerCli.Out(), "To add a manager to this swarm, run the following command:\n\n    docker swarm join --token %s %s\n\n", sw.JoinTokens.Manager, node.ManagerStatus.Addr)
		}
	}

	return nil
}
