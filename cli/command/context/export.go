package context

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/context/store"
	"github.com/spf13/cobra"
)

// ExportOptions are the options used for exporting a context
type ExportOptions struct {
	ContextName string
	Dest        string
}

func newExportCommand(dockerCli command.Cli) *cobra.Command {
	opts := &ExportOptions{}
	cmd := &cobra.Command{
		Use:   "export [OPTIONS] CONTEXT [FILE|-]",
		Short: "Export a context to a tar archive FILE or a tar stream on STDOUT.",
		Args:  cli.RequiresRangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ContextName = args[0]
			if len(args) == 2 {
				opts.Dest = args[1]
			} else {
				opts.Dest = opts.ContextName + ".dockercontext"
			}
			return RunExport(dockerCli, opts)
		},
	}

	flags := cmd.Flags()
	flags.Bool("kubeconfig", false, "Export as a kubeconfig file")
	flags.MarkDeprecated("kubeconfig", "option will be ignored")
	flags.SetAnnotation("kubeconfig", "kubernetes", nil)
	flags.SetAnnotation("kubeconfig", "deprecated", nil)
	return cmd
}

func writeTo(dockerCli command.Cli, reader io.Reader, dest string) error {
	var writer io.Writer
	var printDest bool
	if dest == "-" {
		if dockerCli.Out().IsTerminal() {
			return errors.New("cowardly refusing to export to a terminal, please specify a file path")
		}
		writer = dockerCli.Out()
	} else {
		f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return err
		}
		defer f.Close()
		writer = f
		printDest = true
	}
	if _, err := io.Copy(writer, reader); err != nil {
		return err
	}
	if printDest {
		fmt.Fprintf(dockerCli.Err(), "Written file %q\n", dest)
	}
	return nil
}

// RunExport exports a Docker context
func RunExport(dockerCli command.Cli, opts *ExportOptions) error {
	if err := store.ValidateContextName(opts.ContextName); err != nil && opts.ContextName != command.DefaultContextName {
		return err
	}
	reader := store.Export(opts.ContextName, dockerCli.ContextStore())
	defer reader.Close()
	return writeTo(dockerCli, reader, opts.Dest)
}
