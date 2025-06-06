package service

import (
	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/spf13/cobra"
)

// NewServiceCommand returns a cobra command for `service` subcommands
func NewServiceCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage Swarm services",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
		Annotations: map[string]string{
			"version": "1.24",
			"swarm":   "manager",
		},
	}
	cmd.AddCommand(
		newCreateCommand(dockerCli),
		newInspectCommand(dockerCli),
		newPsCommand(dockerCli),
		newListCommand(dockerCli),
		newRemoveCommand(dockerCli),
		newScaleCommand(dockerCli),
		newUpdateCommand(dockerCli),
		newLogsCommand(dockerCli),
		newRollbackCommand(dockerCli),
	)
	return cmd
}

// CompletionFn offers completion for swarm services
func CompletionFn(dockerCli command.Cli) completion.ValidArgsFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, err := dockerCli.Client().ServiceList(cmd.Context(), types.ServiceListOptions{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		var names []string
		for _, service := range list {
			names = append(names, service.ID)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}
}
