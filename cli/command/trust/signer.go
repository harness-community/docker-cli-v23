package trust

import (
	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/spf13/cobra"
)

// newTrustSignerCommand returns a cobra command for `trust signer` subcommands
func newTrustSignerCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signer",
		Short: "Manage entities who can sign Docker images",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
	}
	cmd.AddCommand(
		newSignerAddCommand(dockerCli),
		newSignerRemoveCommand(dockerCli),
	)
	return cmd
}
