package system

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/opts"
	"github.com/harness-community/docker-cli-v23/templates"
	"github.com/harness-community/docker-v23/api/types"
	eventtypes "github.com/harness-community/docker-v23/api/types/events"
	"github.com/spf13/cobra"
)

type eventsOptions struct {
	since  string
	until  string
	filter opts.FilterOpt
	format string
}

// NewEventsCommand creates a new cobra.Command for `docker events`
func NewEventsCommand(dockerCli command.Cli) *cobra.Command {
	options := eventsOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "events [OPTIONS]",
		Short: "Get real time events from the server",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEvents(dockerCli, &options)
		},
		Annotations: map[string]string{
			"aliases": "docker system events, docker events",
		},
		ValidArgsFunction: completion.NoComplete,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.since, "since", "", "Show all events created since timestamp")
	flags.StringVar(&options.until, "until", "", "Stream events until this timestamp")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")
	flags.StringVar(&options.format, "format", "", "Format the output using the given Go template")

	return cmd
}

func runEvents(dockerCli command.Cli, options *eventsOptions) error {
	tmpl, err := makeTemplate(options.format)
	if err != nil {
		return cli.StatusError{
			StatusCode: 64,
			Status:     "Error parsing format: " + err.Error(),
		}
	}
	eventOptions := types.EventsOptions{
		Since:   options.since,
		Until:   options.until,
		Filters: options.filter.Value(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	events, errs := dockerCli.Client().Events(ctx, eventOptions)
	defer cancel()

	out := dockerCli.Out()

	for {
		select {
		case event := <-events:
			if err := handleEvent(out, event, tmpl); err != nil {
				return err
			}
		case err := <-errs:
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func handleEvent(out io.Writer, event eventtypes.Message, tmpl *template.Template) error {
	if tmpl == nil {
		return prettyPrintEvent(out, event)
	}

	return formatEvent(out, event, tmpl)
}

func makeTemplate(format string) (*template.Template, error) {
	if format == "" {
		return nil, nil
	}
	tmpl, err := templates.Parse(format)
	if err != nil {
		return tmpl, err
	}
	// we execute the template for an empty message, so as to validate
	// a bad template like "{{.badFieldString}}"
	return tmpl, tmpl.Execute(io.Discard, &eventtypes.Message{})
}

// rfc3339NanoFixed is similar to time.RFC3339Nano, except it pads nanoseconds
// zeros to maintain a fixed number of characters
const rfc3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"

// prettyPrintEvent prints all types of event information.
// Each output includes the event type, actor id, name and action.
// Actor attributes are printed at the end if the actor has any.
func prettyPrintEvent(out io.Writer, event eventtypes.Message) error {
	if event.TimeNano != 0 {
		fmt.Fprintf(out, "%s ", time.Unix(0, event.TimeNano).Format(rfc3339NanoFixed))
	} else if event.Time != 0 {
		fmt.Fprintf(out, "%s ", time.Unix(event.Time, 0).Format(rfc3339NanoFixed))
	}

	fmt.Fprintf(out, "%s %s %s", event.Type, event.Action, event.Actor.ID)

	if len(event.Actor.Attributes) > 0 {
		var attrs []string
		var keys []string
		for k := range event.Actor.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := event.Actor.Attributes[k]
			attrs = append(attrs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(out, " (%s)", strings.Join(attrs, ", "))
	}
	fmt.Fprint(out, "\n")
	return nil
}

func formatEvent(out io.Writer, event eventtypes.Message, tmpl *template.Template) error {
	defer out.Write([]byte{'\n'})
	return tmpl.Execute(out, event)
}
