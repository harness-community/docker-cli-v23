package stack

import (
	"github.com/DevanshMathur19/docker-cli-v23/cli"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/stack/options"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/stack/swarm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newRemoveCommand(dockerCli command.Cli) *cobra.Command {
	var opts options.Remove

	cmd := &cobra.Command{
		Use:     "rm [OPTIONS] STACK [STACK...]",
		Aliases: []string{"remove", "down"},
		Short:   "Remove one or more stacks",
		Args:    cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Namespaces = args
			if err := validateStackNames(opts.Namespaces); err != nil {
				return err
			}
			return RunRemove(dockerCli, cmd.Flags(), opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeNames(dockerCli)(cmd, args, toComplete)
		},
	}
	return cmd
}

// RunRemove performs a stack remove against the specified swarm cluster
func RunRemove(dockerCli command.Cli, flags *pflag.FlagSet, opts options.Remove) error {
	return swarm.RunRemove(dockerCli, opts)
}
