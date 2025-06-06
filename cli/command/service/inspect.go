package service

import (
	"context"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	flagsHelper "github.com/harness-community/docker-cli-v23/cli/flags"
	"github.com/harness-community/docker-v23/api/types"
	apiclient "github.com/harness-community/docker-v23/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type inspectOptions struct {
	refs   []string
	format string
	pretty bool
}

func newInspectCommand(dockerCli command.Cli) *cobra.Command {
	var opts inspectOptions

	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] SERVICE [SERVICE...]",
		Short: "Display detailed information on one or more services",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.refs = args

			if opts.pretty && len(opts.format) > 0 {
				return errors.Errorf("--format is incompatible with human friendly format")
			}
			return runInspect(dockerCli, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return CompletionFn(dockerCli)(cmd, args, toComplete)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.format, "format", "f", "", flagsHelper.InspectFormatHelp)
	flags.BoolVar(&opts.pretty, "pretty", false, "Print the information in a human friendly format")
	return cmd
}

func runInspect(dockerCli command.Cli, opts inspectOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	if opts.pretty {
		opts.format = "pretty"
	}

	getRef := func(ref string) (interface{}, []byte, error) {
		// Service inspect shows defaults values in empty fields.
		service, _, err := client.ServiceInspectWithRaw(ctx, ref, types.ServiceInspectOptions{InsertDefaults: true})
		if err == nil || !apiclient.IsErrNotFound(err) {
			return service, nil, err
		}
		return nil, nil, errors.Errorf("Error: no such service: %s", ref)
	}

	getNetwork := func(ref string) (interface{}, []byte, error) {
		network, _, err := client.NetworkInspectWithRaw(ctx, ref, types.NetworkInspectOptions{Scope: "swarm"})
		if err == nil || !apiclient.IsErrNotFound(err) {
			return network, nil, err
		}
		return nil, nil, errors.Errorf("Error: no such network: %s", ref)
	}

	f := opts.format
	if len(f) == 0 {
		f = "raw"
		if len(dockerCli.ConfigFile().ServiceInspectFormat) > 0 {
			f = dockerCli.ConfigFile().ServiceInspectFormat
		}
	}

	// check if the user is trying to apply a template to the pretty format, which
	// is not supported
	if strings.HasPrefix(f, "pretty") && f != "pretty" {
		return errors.Errorf("Cannot supply extra formatting options to the pretty template")
	}

	serviceCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewFormat(f),
	}

	if err := InspectFormatWrite(serviceCtx, opts.refs, getRef, getNetwork); err != nil {
		return cli.StatusError{StatusCode: 1, Status: err.Error()}
	}
	return nil
}
