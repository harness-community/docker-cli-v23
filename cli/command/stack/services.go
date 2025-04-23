package stack

import (
	"fmt"
	"sort"

	"github.com/DevanshMathur19/docker-cli-v23/cli"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/service"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/stack/formatter"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/stack/options"
	"github.com/DevanshMathur19/docker-cli-v23/cli/command/stack/swarm"
	flagsHelper "github.com/DevanshMathur19/docker-cli-v23/cli/flags"
	cliopts "github.com/DevanshMathur19/docker-cli-v23/opts"
	swarmtypes "github.com/DevanshMathur19/docker-v23/api/types/swarm"
	"github.com/fvbommel/sortorder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newServicesCommand(dockerCli command.Cli) *cobra.Command {
	opts := options.Services{Filter: cliopts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "services [OPTIONS] STACK",
		Short: "List the services in the stack",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Namespace = args[0]
			if err := validateStackName(opts.Namespace); err != nil {
				return err
			}
			return RunServices(dockerCli, cmd.Flags(), opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeNames(dockerCli)(cmd, args, toComplete)
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&opts.Quiet, "quiet", "q", false, "Only display IDs")
	flags.StringVar(&opts.Format, "format", "", flagsHelper.FormatHelp)
	flags.VarP(&opts.Filter, "filter", "f", "Filter output based on conditions provided")
	return cmd
}

// RunServices performs a stack services against the specified swarm cluster
func RunServices(dockerCli command.Cli, flags *pflag.FlagSet, opts options.Services) error {
	services, err := GetServices(dockerCli, flags, opts)
	if err != nil {
		return err
	}
	return formatWrite(dockerCli, services, opts)
}

// GetServices returns the services for the specified swarm cluster
func GetServices(dockerCli command.Cli, flags *pflag.FlagSet, opts options.Services) ([]swarmtypes.Service, error) {
	return swarm.GetServices(dockerCli, opts)
}

func formatWrite(dockerCli command.Cli, services []swarmtypes.Service, opts options.Services) error {
	// if no services in the stack, print message and exit 0
	if len(services) == 0 {
		_, _ = fmt.Fprintf(dockerCli.Err(), "Nothing found in stack: %s\n", opts.Namespace)
		return nil
	}
	sort.Slice(services, func(i, j int) bool {
		return sortorder.NaturalLess(services[i].Spec.Name, services[j].Spec.Name)
	})

	format := opts.Format
	if len(format) == 0 {
		if len(dockerCli.ConfigFile().ServicesFormat) > 0 && !opts.Quiet {
			format = dockerCli.ConfigFile().ServicesFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	servicesCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: service.NewListFormat(format, opts.Quiet),
	}
	return service.ListFormatWrite(servicesCtx, services)
}
