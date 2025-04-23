package trust

import (
	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/spf13/cobra"
)

// NewTrustCommand returns a cobra command for `trust` subcommands
func NewTrustCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trust",
		Short: "Manage trust on Docker images",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
	}
	cmd.AddCommand(
		newRevokeCommand(dockerCli),
		newSignCommand(dockerCli),
		newTrustKeyCommand(dockerCli),
		newTrustSignerCommand(dockerCli),
		newInspectCommand(dockerCli),
	)
	return cmd
}
