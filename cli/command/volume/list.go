package volume

import (
	"context"
	"sort"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	flagsHelper "github.com/harness-community/docker-cli-v23/cli/flags"
	"github.com/harness-community/docker-cli-v23/opts"
	"github.com/fvbommel/sortorder"
	"github.com/spf13/cobra"
)

const (
	clusterTableFormat = "table {{.Name}}\t{{.Group}}\t{{.Driver}}\t{{.Availability}}\t{{.Status}}"
)

type listOptions struct {
	quiet   bool
	format  string
	cluster bool
	filter  opts.FilterOpt
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	options := listOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List volumes",
		Args:    cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, options)
		},
		ValidArgsFunction: completion.NoComplete,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display volume names")
	flags.StringVar(&options.format, "format", "", flagsHelper.FormatHelp)
	flags.VarP(&options.filter, "filter", "f", `Provide filter values (e.g. "dangling=true")`)
	flags.BoolVar(&options.cluster, "cluster", false, "Display only cluster volumes, and use cluster volume list formatting")
	flags.SetAnnotation("cluster", "version", []string{"1.42"})
	flags.SetAnnotation("cluster", "swarm", []string{"manager"})

	return cmd
}

func runList(dockerCli command.Cli, options listOptions) error {
	client := dockerCli.Client()
	volumes, err := client.VolumeList(context.Background(), options.filter.Value())
	if err != nil {
		return err
	}

	format := options.format
	if len(format) == 0 && !options.cluster {
		if len(dockerCli.ConfigFile().VolumesFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().VolumesFormat
		} else {
			format = formatter.TableFormatKey
		}
	} else if options.cluster {
		// TODO(dperny): write server-side filter for cluster volumes. For this
		// proof of concept, we'll just filter out non-cluster volumes here

		// trick for filtering in place
		n := 0
		for _, volume := range volumes.Volumes {
			if volume.ClusterVolume != nil {
				volumes.Volumes[n] = volume
				n++
			}
		}
		volumes.Volumes = volumes.Volumes[:n]
		if !options.quiet {
			format = clusterTableFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	sort.Slice(volumes.Volumes, func(i, j int) bool {
		return sortorder.NaturalLess(volumes.Volumes[i].Name, volumes.Volumes[j].Name)
	})

	volumeCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: formatter.NewVolumeFormat(format, options.quiet),
	}
	return formatter.VolumeWrite(volumeCtx, volumes.Volumes)
}
