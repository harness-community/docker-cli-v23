package volume

import (
	"context"
	"fmt"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/opts"
	"github.com/harness-community/docker-v23/api/types/versions"
	"github.com/harness-community/docker-v23/errdefs"
	units "github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type pruneOptions struct {
	all    bool
	force  bool
	filter opts.FilterOpt
}

// NewPruneCommand returns a new cobra prune command for volumes
func NewPruneCommand(dockerCli command.Cli) *cobra.Command {
	options := pruneOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "prune [OPTIONS]",
		Short: "Remove all unused local volumes",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceReclaimed, output, err := runPrune(dockerCli, options)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Fprintln(dockerCli.Out(), output)
			}
			fmt.Fprintln(dockerCli.Out(), "Total reclaimed space:", units.HumanSize(float64(spaceReclaimed)))
			return nil
		},
		Annotations:       map[string]string{"version": "1.25"},
		ValidArgsFunction: completion.NoComplete,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.all, "all", "a", false, "Remove all unused volumes, not just anonymous ones")
	flags.SetAnnotation("all", "version", []string{"1.42"})
	flags.BoolVarP(&options.force, "force", "f", false, "Do not prompt for confirmation")
	flags.Var(&options.filter, "filter", `Provide filter values (e.g. "label=<label>")`)

	return cmd
}

const (
	unusedVolumesWarning = `WARNING! This will remove anonymous local volumes not used by at least one container.
Are you sure you want to continue?`
	allVolumesWarning = `WARNING! This will remove all local volumes not used by at least one container.
Are you sure you want to continue?`
)

func runPrune(dockerCli command.Cli, options pruneOptions) (spaceReclaimed uint64, output string, err error) {
	pruneFilters := command.PruneFilters(dockerCli, options.filter.Value())

	warning := unusedVolumesWarning
	if versions.GreaterThanOrEqualTo(dockerCli.CurrentVersion(), "1.42") {
		if options.all {
			if pruneFilters.Contains("all") {
				return 0, "", errdefs.InvalidParameter(errors.New("conflicting options: cannot specify both --all and --filter all=1"))
			}
			pruneFilters.Add("all", "true")
			warning = allVolumesWarning
		}
	} else {
		// API < v1.42 removes all volumes (anonymous and named) by default.
		warning = allVolumesWarning
	}
	if !options.force && !command.PromptForConfirmation(dockerCli.In(), dockerCli.Out(), warning) {
		return 0, "", nil
	}

	report, err := dockerCli.Client().VolumesPrune(context.Background(), pruneFilters)
	if err != nil {
		return 0, "", err
	}

	if len(report.VolumesDeleted) > 0 {
		output = "Deleted Volumes:\n"
		for _, id := range report.VolumesDeleted {
			output += id + "\n"
		}
		spaceReclaimed = report.SpaceReclaimed
	}

	return spaceReclaimed, output, nil
}

// RunPrune calls the Volume Prune API
// This returns the amount of space reclaimed and a detailed output string
func RunPrune(dockerCli command.Cli, all bool, filter opts.FilterOpt) (uint64, string, error) {
	return runPrune(dockerCli, pruneOptions{force: true, filter: filter})
}
