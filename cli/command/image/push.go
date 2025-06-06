package image

import (
	"context"
	"fmt"
	"io"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/streams"
	"github.com/docker/distribution/reference"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/pkg/jsonmessage"
	"github.com/harness-community/docker-v23/registry"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type pushOptions struct {
	all       bool
	remote    string
	untrusted bool
	quiet     bool
}

// NewPushCommand creates a new `docker push` command
func NewPushCommand(dockerCli command.Cli) *cobra.Command {
	var opts pushOptions

	cmd := &cobra.Command{
		Use:   "push [OPTIONS] NAME[:TAG]",
		Short: "Upload an image to a registry",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.remote = args[0]
			return RunPush(dockerCli, opts)
		},
		Annotations: map[string]string{
			"category-top": "6",
			"aliases":      "docker image push, docker push",
		},
		ValidArgsFunction: completion.ImageNames(dockerCli),
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.all, "all-tags", "a", false, "Push all tags of an image to the repository")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Suppress verbose output")
	command.AddTrustSigningFlags(flags, &opts.untrusted, dockerCli.ContentTrustEnabled())

	return cmd
}

// RunPush performs a push against the engine based on the specified options
func RunPush(dockerCli command.Cli, opts pushOptions) error {
	ref, err := reference.ParseNormalizedNamed(opts.remote)
	switch {
	case err != nil:
		return err
	case opts.all && !reference.IsNameOnly(ref):
		return errors.New("tag can't be used with --all-tags/-a")
	case !opts.all && reference.IsNameOnly(ref):
		ref = reference.TagNameOnly(ref)
		if tagged, ok := ref.(reference.Tagged); ok && !opts.quiet {
			_, _ = fmt.Fprintf(dockerCli.Out(), "Using default tag: %s\n", tagged.Tag())
		}
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Resolve the Auth config relevant for this server
	authConfig := command.ResolveAuthConfig(ctx, dockerCli, repoInfo.Index)
	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}
	requestPrivilege := command.RegistryAuthenticationPrivilegedFunc(dockerCli, repoInfo.Index, "push")
	options := types.ImagePushOptions{
		All:           opts.all,
		RegistryAuth:  encodedAuth,
		PrivilegeFunc: requestPrivilege,
	}

	responseBody, err := dockerCli.Client().ImagePush(ctx, reference.FamiliarString(ref), options)
	if err != nil {
		return err
	}

	defer responseBody.Close()
	if !opts.untrusted {
		// TODO PushTrustedReference currently doesn't respect `--quiet`
		return PushTrustedReference(dockerCli, repoInfo, ref, authConfig, responseBody)
	}

	if opts.quiet {
		err = jsonmessage.DisplayJSONMessagesToStream(responseBody, streams.NewOut(io.Discard), nil)
		if err == nil {
			fmt.Fprintln(dockerCli.Out(), ref.String())
		}
		return err
	}
	return jsonmessage.DisplayJSONMessagesToStream(responseBody, dockerCli.Out(), nil)
}
