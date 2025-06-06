package node

import (
	"context"
	"sort"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	flagsHelper "github.com/harness-community/docker-cli-v23/cli/flags"
	"github.com/harness-community/docker-cli-v23/opts"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/fvbommel/sortorder"
	"github.com/spf13/cobra"
)

type listOptions struct {
	quiet  bool
	format string
	filter opts.FilterOpt
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	options := listOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List nodes in the swarm",
		Args:    cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, options)
		},
		ValidArgsFunction: completion.NoComplete,
	}
	flags := cmd.Flags()
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display IDs")
	flags.StringVar(&options.format, "format", "", flagsHelper.FormatHelp)
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")

	return cmd
}

func runList(dockerCli command.Cli, options listOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	nodes, err := client.NodeList(
		ctx,
		types.NodeListOptions{Filters: options.filter.Value()})
	if err != nil {
		return err
	}

	info := types.Info{}
	if len(nodes) > 0 && !options.quiet {
		// only non-empty nodes and not quiet, should we call /info api
		info, err = client.Info(ctx)
		if err != nil {
			return err
		}
	}

	format := options.format
	if len(format) == 0 {
		format = formatter.TableFormatKey
		if len(dockerCli.ConfigFile().NodesFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().NodesFormat
		}
	}

	nodesCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewFormat(format, options.quiet),
	}
	sort.Slice(nodes, func(i, j int) bool {
		return sortorder.NaturalLess(nodes[i].Description.Hostname, nodes[j].Description.Hostname)
	})
	return FormatWrite(nodesCtx, nodes, info)
}
