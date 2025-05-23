package context

import (
	"bytes"
	"fmt"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter/tabwriter"
	"github.com/harness-community/docker-cli-v23/cli/context/docker"
	"github.com/harness-community/docker-cli-v23/cli/context/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// UpdateOptions are the options used to update a context
type UpdateOptions struct {
	Name        string
	Description string
	Docker      map[string]string
}

func longUpdateDescription() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("Update a context\n\nDocker endpoint config:\n\n")
	tw := tabwriter.NewWriter(buf, 20, 1, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION")
	for _, d := range dockerConfigKeysDescriptions {
		fmt.Fprintf(tw, "%s\t%s\n", d.name, d.description)
	}
	tw.Flush()
	buf.WriteString("\nExample:\n\n$ docker context update my-context --description \"some description\" --docker \"host=tcp://myserver:2376,ca=~/ca-file,cert=~/cert-file,key=~/key-file\"\n")
	return buf.String()
}

func newUpdateCommand(dockerCli command.Cli) *cobra.Command {
	opts := &UpdateOptions{}
	cmd := &cobra.Command{
		Use:   "update [OPTIONS] CONTEXT",
		Short: "Update a context",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			return RunUpdate(dockerCli, opts)
		},
		Long: longUpdateDescription(),
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.Description, "description", "", "Description of the context")
	flags.String(
		"default-stack-orchestrator", "",
		"Default orchestrator for stack operations to use with this context (swarm|kubernetes|all)",
	)
	flags.SetAnnotation("default-stack-orchestrator", "deprecated", nil)
	flags.MarkDeprecated("default-stack-orchestrator", "option will be ignored")
	flags.StringToStringVar(&opts.Docker, "docker", nil, "set the docker endpoint")
	flags.StringToString("kubernetes", nil, "set the kubernetes endpoint")
	flags.SetAnnotation("kubernetes", "kubernetes", nil)
	flags.SetAnnotation("kubernetes", "deprecated", nil)
	flags.MarkDeprecated("kubernetes", "option will be ignored")
	return cmd
}

// RunUpdate updates a Docker context
func RunUpdate(cli command.Cli, o *UpdateOptions) error {
	if err := store.ValidateContextName(o.Name); err != nil {
		return err
	}
	s := cli.ContextStore()
	c, err := s.GetMetadata(o.Name)
	if err != nil {
		return err
	}
	dockerContext, err := command.GetDockerContext(c)
	if err != nil {
		return err
	}
	if o.Description != "" {
		dockerContext.Description = o.Description
	}

	c.Metadata = dockerContext

	tlsDataToReset := make(map[string]*store.EndpointTLSData)

	if o.Docker != nil {
		dockerEP, dockerTLS, err := getDockerEndpointMetadataAndTLS(cli, o.Docker)
		if err != nil {
			return errors.Wrap(err, "unable to create docker endpoint config")
		}
		c.Endpoints[docker.DockerEndpoint] = dockerEP
		tlsDataToReset[docker.DockerEndpoint] = dockerTLS
	}
	if err := validateEndpoints(c); err != nil {
		return err
	}
	if err := s.CreateOrUpdate(c); err != nil {
		return err
	}
	for ep, tlsData := range tlsDataToReset {
		if err := s.ResetEndpointTLSMaterial(o.Name, ep, tlsData); err != nil {
			return err
		}
	}

	fmt.Fprintln(cli.Out(), o.Name)
	fmt.Fprintf(cli.Err(), "Successfully updated context %q\n", o.Name)
	return nil
}

func validateEndpoints(c store.Metadata) error {
	_, err := command.GetDockerContext(c)
	return err
}
