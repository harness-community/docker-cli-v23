package plugin

import (
	"context"
	"sort"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	flagsHelper "github.com/harness-community/docker-cli-v23/cli/flags"
	"github.com/harness-community/docker-cli-v23/opts"
	"github.com/fvbommel/sortorder"
	"github.com/spf13/cobra"
)

type listOptions struct {
	quiet   bool
	noTrunc bool
	format  string
	filter  opts.FilterOpt
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	options := listOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Short:   "List plugins",
		Aliases: []string{"list"},
		Args:    cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, options)
		},
		ValidArgsFunction: completion.NoComplete,
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display plugin IDs")
	flags.BoolVar(&options.noTrunc, "no-trunc", false, "Don't truncate output")
	flags.StringVar(&options.format, "format", "", flagsHelper.FormatHelp)
	flags.VarP(&options.filter, "filter", "f", `Provide filter values (e.g. "enabled=true")`)

	return cmd
}

func runList(dockerCli command.Cli, options listOptions) error {
	plugins, err := dockerCli.Client().PluginList(context.Background(), options.filter.Value())
	if err != nil {
		return err
	}

	sort.Slice(plugins, func(i, j int) bool {
		return sortorder.NaturalLess(plugins[i].Name, plugins[j].Name)
	})

	format := options.format
	if len(format) == 0 {
		if len(dockerCli.ConfigFile().PluginsFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().PluginsFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	pluginsCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewFormat(format, options.quiet),
		Trunc:  !options.noTrunc,
	}
	return FormatWrite(pluginsCtx, plugins)
}
