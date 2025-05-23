package image

import (
	"fmt"
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/filters"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestNewPruneCommandErrors(t *testing.T) {
	testCases := []struct {
		name            string
		args            []string
		expectedError   string
		imagesPruneFunc func(pruneFilter filters.Args) (types.ImagesPruneReport, error)
	}{
		{
			name:          "wrong-args",
			args:          []string{"something"},
			expectedError: "accepts no arguments.",
		},
		{
			name:          "prune-error",
			args:          []string{"--force"},
			expectedError: "something went wrong",
			imagesPruneFunc: func(pruneFilter filters.Args) (types.ImagesPruneReport, error) {
				return types.ImagesPruneReport{}, errors.Errorf("something went wrong")
			},
		},
	}
	for _, tc := range testCases {
		cmd := NewPruneCommand(test.NewFakeCli(&fakeClient{
			imagesPruneFunc: tc.imagesPruneFunc,
		}))
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNewPruneCommandSuccess(t *testing.T) {
	testCases := []struct {
		name            string
		args            []string
		imagesPruneFunc func(pruneFilter filters.Args) (types.ImagesPruneReport, error)
	}{
		{
			name: "all",
			args: []string{"--all"},
			imagesPruneFunc: func(pruneFilter filters.Args) (types.ImagesPruneReport, error) {
				assert.Check(t, is.Equal("false", pruneFilter.Get("dangling")[0]))
				return types.ImagesPruneReport{}, nil
			},
		},
		{
			name: "force-deleted",
			args: []string{"--force"},
			imagesPruneFunc: func(pruneFilter filters.Args) (types.ImagesPruneReport, error) {
				assert.Check(t, is.Equal("true", pruneFilter.Get("dangling")[0]))
				return types.ImagesPruneReport{
					ImagesDeleted:  []types.ImageDeleteResponseItem{{Deleted: "image1"}},
					SpaceReclaimed: 1,
				}, nil
			},
		},
		{
			name: "label-filter",
			args: []string{"--force", "--filter", "label=foobar"},
			imagesPruneFunc: func(pruneFilter filters.Args) (types.ImagesPruneReport, error) {
				assert.Check(t, is.Equal("foobar", pruneFilter.Get("label")[0]))
				return types.ImagesPruneReport{}, nil
			},
		},
		{
			name: "force-untagged",
			args: []string{"--force"},
			imagesPruneFunc: func(pruneFilter filters.Args) (types.ImagesPruneReport, error) {
				assert.Check(t, is.Equal("true", pruneFilter.Get("dangling")[0]))
				return types.ImagesPruneReport{
					ImagesDeleted:  []types.ImageDeleteResponseItem{{Untagged: "image1"}},
					SpaceReclaimed: 2,
				}, nil
			},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{imagesPruneFunc: tc.imagesPruneFunc})
		cmd := NewPruneCommand(cli)
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		assert.NilError(t, err)
		golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("prune-command-success.%s.golden", tc.name))
	}
}
