package swarm

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/streams"
	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newUnlockCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock swarm",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlock(dockerCli)
		},
		Annotations: map[string]string{
			"version": "1.24",
			"swarm":   "manager",
		},
		ValidArgsFunction: completion.NoComplete,
	}

	return cmd
}

func runUnlock(dockerCli command.Cli) error {
	client := dockerCli.Client()
	ctx := context.Background()

	// First see if the node is actually part of a swarm, and if it is actually locked first.
	// If it's in any other state than locked, don't ask for the key.
	info, err := client.Info(ctx)
	if err != nil {
		return err
	}

	switch info.Swarm.LocalNodeState {
	case swarm.LocalNodeStateInactive:
		return errors.New("Error: This node is not part of a swarm")
	case swarm.LocalNodeStateLocked:
		break
	default:
		return errors.New("Error: swarm is not locked")
	}

	key, err := readKey(dockerCli.In(), "Please enter unlock key: ")
	if err != nil {
		return err
	}
	req := swarm.UnlockRequest{
		UnlockKey: key,
	}

	return client.SwarmUnlock(ctx, req)
}

func readKey(in *streams.In, prompt string) (string, error) {
	if in.IsTerminal() {
		fmt.Print(prompt)
		dt, err := term.ReadPassword(int(in.FD()))
		fmt.Println()
		return string(dt), err
	}
	key, err := bufio.NewReader(in).ReadString('\n')
	if err == io.EOF {
		err = nil
	}
	return strings.TrimSpace(key), err
}
