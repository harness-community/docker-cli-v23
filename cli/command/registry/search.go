package registry

import (
	"context"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	"github.com/harness-community/docker-cli-v23/opts"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/registry"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	format  string
	term    string
	noTrunc bool
	limit   int
	filter  opts.FilterOpt
}

// NewSearchCommand creates a new `docker search` command
func NewSearchCommand(dockerCli command.Cli) *cobra.Command {
	options := searchOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "search [OPTIONS] TERM",
		Short: "Search Docker Hub for images",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.term = args[0]
			return runSearch(dockerCli, options)
		},
		Annotations: map[string]string{
			"category-top": "10",
		},
	}

	flags := cmd.Flags()

	flags.BoolVar(&options.noTrunc, "no-trunc", false, "Don't truncate output")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")
	flags.IntVar(&options.limit, "limit", 0, "Max number of search results")
	flags.StringVar(&options.format, "format", "", "Pretty-print search using a Go template")

	return cmd
}

func runSearch(dockerCli command.Cli, options searchOptions) error {
	indexInfo, err := registry.ParseSearchIndexInfo(options.term)
	if err != nil {
		return err
	}

	ctx := context.Background()

	authConfig := command.ResolveAuthConfig(ctx, dockerCli, indexInfo)
	requestPrivilege := command.RegistryAuthenticationPrivilegedFunc(dockerCli, indexInfo, "search")

	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}

	searchOptions := types.ImageSearchOptions{
		RegistryAuth:  encodedAuth,
		PrivilegeFunc: requestPrivilege,
		Filters:       options.filter.Value(),
		Limit:         options.limit,
	}

	clnt := dockerCli.Client()

	results, err := clnt.ImageSearch(ctx, options.term, searchOptions)
	if err != nil {
		return err
	}

	searchCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewSearchFormat(options.format),
		Trunc:  !options.noTrunc,
	}
	return SearchWrite(searchCtx, results)
}
