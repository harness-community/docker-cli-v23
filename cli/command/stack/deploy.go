package stack

import (
	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/stack/loader"
	"github.com/harness-community/docker-cli-v23/cli/command/stack/options"
	"github.com/harness-community/docker-cli-v23/cli/command/stack/swarm"
	composetypes "github.com/harness-community/docker-cli-v23/cli/compose/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newDeployCommand(dockerCli command.Cli) *cobra.Command {
	var opts options.Deploy

	cmd := &cobra.Command{
		Use:     "deploy [OPTIONS] STACK",
		Aliases: []string{"up"},
		Short:   "Deploy a new stack or update an existing stack",
		Args:    cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Namespace = args[0]
			if err := validateStackName(opts.Namespace); err != nil {
				return err
			}
			config, err := loader.LoadComposefile(dockerCli, opts)
			if err != nil {
				return err
			}
			return RunDeploy(dockerCli, cmd.Flags(), config, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeNames(dockerCli)(cmd, args, toComplete)
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opts.Composefiles, "compose-file", "c", []string{}, `Path to a Compose file, or "-" to read from stdin`)
	flags.SetAnnotation("compose-file", "version", []string{"1.25"})
	flags.BoolVar(&opts.SendRegistryAuth, "with-registry-auth", false, "Send registry authentication details to Swarm agents")
	flags.BoolVar(&opts.Prune, "prune", false, "Prune services that are no longer referenced")
	flags.SetAnnotation("prune", "version", []string{"1.27"})
	flags.StringVar(&opts.ResolveImage, "resolve-image", swarm.ResolveImageAlways,
		`Query the registry to resolve image digest and supported platforms ("`+swarm.ResolveImageAlways+`", "`+swarm.ResolveImageChanged+`", "`+swarm.ResolveImageNever+`")`)
	flags.SetAnnotation("resolve-image", "version", []string{"1.30"})
	return cmd
}

// RunDeploy performs a stack deploy against the specified swarm cluster
func RunDeploy(dockerCli command.Cli, flags *pflag.FlagSet, config *composetypes.Config, opts options.Deploy) error {
	return swarm.RunDeploy(dockerCli, opts, config)
}
