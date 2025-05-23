package context

import (
	"fmt"
	"os"

	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/context/store"
	"github.com/harness-community/docker-v23/client"
	"github.com/spf13/cobra"
)

func newUseCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use CONTEXT",
		Short: "Set the current docker context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return RunUse(dockerCli, name)
		},
	}
	return cmd
}

// RunUse set the current Docker context
func RunUse(dockerCli command.Cli, name string) error {
	// configValue uses an empty string for "default"
	var configValue string
	if name != command.DefaultContextName {
		if err := store.ValidateContextName(name); err != nil {
			return err
		}
		if _, err := dockerCli.ContextStore().GetMetadata(name); err != nil {
			return err
		}
		configValue = name
	}
	dockerConfig := dockerCli.ConfigFile()
	// Avoid updating the config-file if nothing changed. This also prevents
	// creating the file and config-directory if the default is used and
	// no config-file existed yet.
	if dockerConfig.CurrentContext != configValue {
		dockerConfig.CurrentContext = configValue
		if err := dockerConfig.Save(); err != nil {
			return err
		}
	}
	fmt.Fprintln(dockerCli.Out(), name)
	fmt.Fprintf(dockerCli.Err(), "Current context is now %q\n", name)
	if name != command.DefaultContextName && os.Getenv(client.EnvOverrideHost) != "" {
		fmt.Fprintf(dockerCli.Err(), "Warning: %[1]s environment variable overrides the active context. "+
			"To use %[2]q, either set the global --context flag, or unset %[1]s environment variable.\n", client.EnvOverrideHost, name)
	}
	return nil
}
