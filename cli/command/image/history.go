package image

import (
	"context"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	flagsHelper "github.com/harness-community/docker-cli-v23/cli/flags"
	"github.com/spf13/cobra"
)

type historyOptions struct {
	image string

	human   bool
	quiet   bool
	noTrunc bool
	format  string
}

// NewHistoryCommand creates a new `docker history` command
func NewHistoryCommand(dockerCli command.Cli) *cobra.Command {
	var opts historyOptions

	cmd := &cobra.Command{
		Use:   "history [OPTIONS] IMAGE",
		Short: "Show the history of an image",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.image = args[0]
			return runHistory(dockerCli, opts)
		},
		Annotations: map[string]string{
			"aliases": "docker image history, docker history",
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.human, "human", "H", true, "Print sizes and dates in human readable format")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Only show image IDs")
	flags.BoolVar(&opts.noTrunc, "no-trunc", false, "Don't truncate output")
	flags.StringVar(&opts.format, "format", "", flagsHelper.FormatHelp)

	return cmd
}

func runHistory(dockerCli command.Cli, opts historyOptions) error {
	ctx := context.Background()

	history, err := dockerCli.Client().ImageHistory(ctx, opts.image)
	if err != nil {
		return err
	}

	format := opts.format
	if len(format) == 0 {
		format = formatter.TableFormatKey
	}

	historyCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewHistoryFormat(format, opts.quiet, opts.human),
		Trunc:  !opts.noTrunc,
	}
	return HistoryWrite(historyCtx, opts.human, history)
}
