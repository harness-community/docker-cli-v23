package swarm

import (
	"io"
	"strings"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestSwarmJoinErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		swarmJoinFunc func() error
		infoFunc      func() (types.Info, error)
		expectedError string
	}{
		{
			name:          "not-enough-args",
			expectedError: "requires exactly 1 argument",
		},
		{
			name:          "too-many-args",
			args:          []string{"remote1", "remote2"},
			expectedError: "requires exactly 1 argument",
		},
		{
			name: "join-failed",
			args: []string{"remote"},
			swarmJoinFunc: func() error {
				return errors.Errorf("error joining the swarm")
			},
			expectedError: "error joining the swarm",
		},
		{
			name: "join-failed-on-init",
			args: []string{"remote"},
			infoFunc: func() (types.Info, error) {
				return types.Info{}, errors.Errorf("error asking for node info")
			},
			expectedError: "error asking for node info",
		},
	}
	for _, tc := range testCases {
		cmd := newJoinCommand(
			test.NewFakeCli(&fakeClient{
				swarmJoinFunc: tc.swarmJoinFunc,
				infoFunc:      tc.infoFunc,
			}))
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestSwarmJoin(t *testing.T) {
	testCases := []struct {
		name     string
		infoFunc func() (types.Info, error)
		expected string
	}{
		{
			name: "join-as-manager",
			infoFunc: func() (types.Info, error) {
				return types.Info{
					Swarm: swarm.Info{
						ControlAvailable: true,
					},
				}, nil
			},
			expected: "This node joined a swarm as a manager.",
		},
		{
			name: "join-as-worker",
			infoFunc: func() (types.Info, error) {
				return types.Info{
					Swarm: swarm.Info{
						ControlAvailable: false,
					},
				}, nil
			},
			expected: "This node joined a swarm as a worker.",
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			infoFunc: tc.infoFunc,
		})
		cmd := newJoinCommand(cli)
		cmd.SetArgs([]string{"remote"})
		assert.NilError(t, cmd.Execute())
		assert.Check(t, is.Equal(strings.TrimSpace(cli.OutBuffer().String()), tc.expected))
	}
}
