package image

import (
	"context"

	"github.com/DevanshMathur19/cli-v23/cli"
	"github.com/DevanshMathur19/cli-v23/cli/command"
	"github.com/DevanshMathur19/cli-v23/cli/command/inspect"
	flagsHelper "github.com/DevanshMathur19/cli-v23/cli/flags"
	"github.com/spf13/cobra"
)

type inspectOptions struct {
	format string
	refs   []string
}

// newInspectCommand creates a new cobra.Command for `docker image inspect`
func newInspectCommand(dockerCli command.Cli) *cobra.Command {
	var opts inspectOptions

	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] IMAGE [IMAGE...]",
		Short: "Display detailed information on one or more images",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.refs = args
			return runInspect(dockerCli, opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.format, "format", "f", "", flagsHelper.InspectFormatHelp)
	return cmd
}

func runInspect(dockerCli command.Cli, opts inspectOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	getRefFunc := func(ref string) (interface{}, []byte, error) {
		return client.ImageInspectWithRaw(ctx, ref)
	}
	return inspect.Inspect(dockerCli.Out(), opts.refs, opts.format, getRefFunc)
}
