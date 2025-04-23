package network

import (
	"context"

	"github.com/DevanshMathur19/docker-cli-v23/cli"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/completion"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/inspect"
	flagsHelper "github.com/DevanshMathur19/docker-cli-v23/cli/flags"
	"github.com/DevanshMathur19/docker-v23/api/types"
	"github.com/spf13/cobra"
)

type inspectOptions struct {
	format  string
	names   []string
	verbose bool
}

func newInspectCommand(dockerCli command.Cli) *cobra.Command {
	var opts inspectOptions

	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] NETWORK [NETWORK...]",
		Short: "Display detailed information on one or more networks",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.names = args
			return runInspect(dockerCli, opts)
		},
		ValidArgsFunction: completion.NetworkNames(dockerCli),
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "", flagsHelper.InspectFormatHelp)
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "Verbose output for diagnostics")

	return cmd
}

func runInspect(dockerCli command.Cli, opts inspectOptions) error {
	client := dockerCli.Client()

	ctx := context.Background()

	getNetFunc := func(name string) (interface{}, []byte, error) {
		return client.NetworkInspectWithRaw(ctx, name, types.NetworkInspectOptions{Verbose: opts.verbose})
	}

	return inspect.Inspect(dockerCli.Out(), opts.names, opts.format, getNetFunc)
}
