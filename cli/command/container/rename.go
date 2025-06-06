package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type renameOptions struct {
	oldName string
	newName string
}

// NewRenameCommand creates a new cobra.Command for `docker rename`
func NewRenameCommand(dockerCli command.Cli) *cobra.Command {
	var opts renameOptions

	cmd := &cobra.Command{
		Use:   "rename CONTAINER NEW_NAME",
		Short: "Rename a container",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.oldName = args[0]
			opts.newName = args[1]
			return runRename(dockerCli, &opts)
		},
		Annotations: map[string]string{
			"aliases": "docker container rename, docker rename",
		},
		ValidArgsFunction: completion.ContainerNames(dockerCli, true),
	}
	return cmd
}

func runRename(dockerCli command.Cli, opts *renameOptions) error {
	ctx := context.Background()

	oldName := strings.TrimSpace(opts.oldName)
	newName := strings.TrimSpace(opts.newName)

	if oldName == "" || newName == "" {
		return errors.New("Error: Neither old nor new names may be empty")
	}

	if err := dockerCli.Client().ContainerRename(ctx, oldName, newName); err != nil {
		fmt.Fprintln(dockerCli.Err(), err)
		return errors.Errorf("Error: failed to rename container named %s", oldName)
	}
	return nil
}
