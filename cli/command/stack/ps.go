package stack

import (
	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/stack/options"
	"github.com/harness-community/docker-cli-v23/cli/command/stack/swarm"
	flagsHelper "github.com/harness-community/docker-cli-v23/cli/flags"
	cliopts "github.com/harness-community/docker-cli-v23/opts"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newPsCommand(dockerCli command.Cli) *cobra.Command {
	opts := options.PS{Filter: cliopts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "ps [OPTIONS] STACK",
		Short: "List the tasks in the stack",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Namespace = args[0]
			if err := validateStackName(opts.Namespace); err != nil {
				return err
			}
			return RunPs(dockerCli, cmd.Flags(), opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeNames(dockerCli)(cmd, args, toComplete)
		},
	}
	flags := cmd.Flags()
	flags.BoolVar(&opts.NoTrunc, "no-trunc", false, "Do not truncate output")
	flags.BoolVar(&opts.NoResolve, "no-resolve", false, "Do not map IDs to Names")
	flags.VarP(&opts.Filter, "filter", "f", "Filter output based on conditions provided")
	flags.BoolVarP(&opts.Quiet, "quiet", "q", false, "Only display task IDs")
	flags.StringVar(&opts.Format, "format", "", flagsHelper.FormatHelp)
	return cmd
}

// RunPs performs a stack ps against the specified swarm cluster
func RunPs(dockerCli command.Cli, flags *pflag.FlagSet, opts options.PS) error {
	return swarm.RunPS(dockerCli, opts)
}
