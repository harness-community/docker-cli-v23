package image

import (
	"context"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/spf13/cobra"
)

type tagOptions struct {
	image string
	name  string
}

// NewTagCommand creates a new `docker tag` command
func NewTagCommand(dockerCli command.Cli) *cobra.Command {
	var opts tagOptions

	cmd := &cobra.Command{
		Use:   "tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]",
		Short: "Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.image = args[0]
			opts.name = args[1]
			return runTag(dockerCli, opts)
		},
		Annotations: map[string]string{
			"aliases": "docker image tag, docker tag",
		},
		ValidArgsFunction: completion.ImageNames(dockerCli),
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	return cmd
}

func runTag(dockerCli command.Cli, opts tagOptions) error {
	ctx := context.Background()

	return dockerCli.Client().ImageTag(ctx, opts.image, opts.name)
}
