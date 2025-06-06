package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-v23/api/types"
	apiclient "github.com/harness-community/docker-v23/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	force   bool
	noPrune bool
}

// NewRemoveCommand creates a new `docker remove` command
func NewRemoveCommand(dockerCli command.Cli) *cobra.Command {
	var opts removeOptions

	cmd := &cobra.Command{
		Use:   "rmi [OPTIONS] IMAGE [IMAGE...]",
		Short: "Remove one or more images",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(dockerCli, opts, args)
		},
		Annotations: map[string]string{
			"aliases": "docker image rm, docker image remove, docker rmi",
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.force, "force", "f", false, "Force removal of the image")
	flags.BoolVar(&opts.noPrune, "no-prune", false, "Do not delete untagged parents")

	return cmd
}

func newRemoveCommand(dockerCli command.Cli) *cobra.Command {
	cmd := *NewRemoveCommand(dockerCli)
	cmd.Aliases = []string{"rmi", "remove"}
	cmd.Use = "rm [OPTIONS] IMAGE [IMAGE...]"
	return &cmd
}

func runRemove(dockerCli command.Cli, opts removeOptions, images []string) error {
	client := dockerCli.Client()
	ctx := context.Background()

	options := types.ImageRemoveOptions{
		Force:         opts.force,
		PruneChildren: !opts.noPrune,
	}

	var errs []string
	fatalErr := false
	for _, img := range images {
		dels, err := client.ImageRemove(ctx, img, options)
		if err != nil {
			if !apiclient.IsErrNotFound(err) {
				fatalErr = true
			}
			errs = append(errs, err.Error())
		} else {
			for _, del := range dels {
				if del.Deleted != "" {
					fmt.Fprintf(dockerCli.Out(), "Deleted: %s\n", del.Deleted)
				} else {
					fmt.Fprintf(dockerCli.Out(), "Untagged: %s\n", del.Untagged)
				}
			}
		}
	}

	if len(errs) > 0 {
		msg := strings.Join(errs, "\n")
		if !opts.force || fatalErr {
			return errors.New(msg)
		}
		fmt.Fprintln(dockerCli.Err(), msg)
	}
	return nil
}
