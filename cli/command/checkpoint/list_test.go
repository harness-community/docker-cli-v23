package checkpoint

import (
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestCheckpointListErrors(t *testing.T) {
	testCases := []struct {
		args               []string
		checkpointListFunc func(container string, options types.CheckpointListOptions) ([]types.Checkpoint, error)
		expectedError      string
	}{
		{
			args:          []string{},
			expectedError: "requires exactly 1 argument",
		},
		{
			args:          []string{"too", "many", "arguments"},
			expectedError: "requires exactly 1 argument",
		},
		{
			args: []string{"foo"},
			checkpointListFunc: func(container string, options types.CheckpointListOptions) ([]types.Checkpoint, error) {
				return []types.Checkpoint{}, errors.Errorf("error getting checkpoints for container foo")
			},
			expectedError: "error getting checkpoints for container foo",
		},
	}

	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			checkpointListFunc: tc.checkpointListFunc,
		})
		cmd := newListCommand(cli)
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestCheckpointListWithOptions(t *testing.T) {
	var containerID, checkpointDir string
	cli := test.NewFakeCli(&fakeClient{
		checkpointListFunc: func(container string, options types.CheckpointListOptions) ([]types.Checkpoint, error) {
			containerID = container
			checkpointDir = options.CheckpointDir
			return []types.Checkpoint{
				{Name: "checkpoint-foo"},
			}, nil
		},
	})
	cmd := newListCommand(cli)
	cmd.SetArgs([]string{"container-foo"})
	cmd.Flags().Set("checkpoint-dir", "/dir/foo")
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal("container-foo", containerID))
	assert.Check(t, is.Equal("/dir/foo", checkpointDir))
	golden.Assert(t, cli.OutBuffer().String(), "checkpoint-list-with-options.golden")
}
