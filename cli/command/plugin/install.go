package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/image"
	"github.com/docker/distribution/reference"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/pkg/jsonmessage"
	"github.com/harness-community/docker-v23/registry"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type pluginOptions struct {
	remote          string
	localName       string
	grantPerms      bool
	disable         bool
	args            []string
	skipRemoteCheck bool
	untrusted       bool
}

func loadPullFlags(dockerCli command.Cli, opts *pluginOptions, flags *pflag.FlagSet) {
	flags.BoolVar(&opts.grantPerms, "grant-all-permissions", false, "Grant all permissions necessary to run the plugin")
	command.AddTrustVerificationFlags(flags, &opts.untrusted, dockerCli.ContentTrustEnabled())
}

func newInstallCommand(dockerCli command.Cli) *cobra.Command {
	var options pluginOptions
	cmd := &cobra.Command{
		Use:   "install [OPTIONS] PLUGIN [KEY=VALUE...]",
		Short: "Install a plugin",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.remote = args[0]
			if len(args) > 1 {
				options.args = args[1:]
			}
			return runInstall(dockerCli, options)
		},
	}

	flags := cmd.Flags()
	loadPullFlags(dockerCli, &options, flags)
	flags.BoolVar(&options.disable, "disable", false, "Do not enable the plugin on install")
	flags.StringVar(&options.localName, "alias", "", "Local name for plugin")
	return cmd
}

type pluginRegistryService struct {
	registry.Service
}

func (s pluginRegistryService) ResolveRepository(name reference.Named) (*registry.RepositoryInfo, error) {
	repoInfo, err := s.Service.ResolveRepository(name)
	if repoInfo != nil {
		repoInfo.Class = "plugin"
	}
	return repoInfo, err
}

func newRegistryService() (registry.Service, error) {
	svc, err := registry.NewService(registry.ServiceOptions{})
	if err != nil {
		return nil, err
	}
	return pluginRegistryService{Service: svc}, nil
}

func buildPullConfig(ctx context.Context, dockerCli command.Cli, opts pluginOptions, cmdName string) (types.PluginInstallOptions, error) {
	// Names with both tag and digest will be treated by the daemon
	// as a pull by digest with a local name for the tag
	// (if no local name is provided).
	ref, err := reference.ParseNormalizedNamed(opts.remote)
	if err != nil {
		return types.PluginInstallOptions{}, err
	}

	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return types.PluginInstallOptions{}, err
	}

	remote := ref.String()

	_, isCanonical := ref.(reference.Canonical)
	if !opts.untrusted && !isCanonical {
		ref = reference.TagNameOnly(ref)
		nt, ok := ref.(reference.NamedTagged)
		if !ok {
			return types.PluginInstallOptions{}, errors.Errorf("invalid name: %s", ref.String())
		}

		ctx := context.Background()
		svc, err := newRegistryService()
		if err != nil {
			return types.PluginInstallOptions{}, err
		}
		trusted, err := image.TrustedReference(ctx, dockerCli, nt, svc)
		if err != nil {
			return types.PluginInstallOptions{}, err
		}
		remote = reference.FamiliarString(trusted)
	}

	authConfig := command.ResolveAuthConfig(ctx, dockerCli, repoInfo.Index)

	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return types.PluginInstallOptions{}, err
	}
	registryAuthFunc := command.RegistryAuthenticationPrivilegedFunc(dockerCli, repoInfo.Index, cmdName)

	options := types.PluginInstallOptions{
		RegistryAuth:          encodedAuth,
		RemoteRef:             remote,
		Disabled:              opts.disable,
		AcceptAllPermissions:  opts.grantPerms,
		AcceptPermissionsFunc: acceptPrivileges(dockerCli, opts.remote),
		PrivilegeFunc:         registryAuthFunc,
		Args:                  opts.args,
	}
	return options, nil
}

func runInstall(dockerCli command.Cli, opts pluginOptions) error {
	var localName string
	if opts.localName != "" {
		aref, err := reference.ParseNormalizedNamed(opts.localName)
		if err != nil {
			return err
		}
		if _, ok := aref.(reference.Canonical); ok {
			return errors.Errorf("invalid name: %s", opts.localName)
		}
		localName = reference.FamiliarString(reference.TagNameOnly(aref))
	}

	ctx := context.Background()
	options, err := buildPullConfig(ctx, dockerCli, opts, "plugin install")
	if err != nil {
		return err
	}
	responseBody, err := dockerCli.Client().PluginInstall(ctx, localName, options)
	if err != nil {
		if strings.Contains(err.Error(), "(image) when fetching") {
			return errors.New(err.Error() + " - Use \"docker image pull\"")
		}
		return err
	}
	defer responseBody.Close()
	if err := jsonmessage.DisplayJSONMessagesToStream(responseBody, dockerCli.Out(), nil); err != nil {
		return err
	}
	fmt.Fprintf(dockerCli.Out(), "Installed plugin %s\n", opts.remote) // todo: return proper values from the API for this result
	return nil
}

func acceptPrivileges(dockerCli command.Cli, name string) func(privileges types.PluginPrivileges) (bool, error) {
	return func(privileges types.PluginPrivileges) (bool, error) {
		fmt.Fprintf(dockerCli.Out(), "Plugin %q is requesting the following privileges:\n", name)
		for _, privilege := range privileges {
			fmt.Fprintf(dockerCli.Out(), " - %s: %v\n", privilege.Name, privilege.Value)
		}
		return command.PromptForConfirmation(dockerCli.In(), dockerCli.Out(), "Do you grant the above permissions?"), nil
	}
}
