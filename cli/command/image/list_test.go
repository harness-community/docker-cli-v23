package image

import (
	"fmt"
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/config/configfile"
	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestNewImagesCommandErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedError string
		imageListFunc func(options types.ImageListOptions) ([]types.ImageSummary, error)
	}{
		{
			name:          "wrong-args",
			args:          []string{"arg1", "arg2"},
			expectedError: "requires at most 1 argument.",
		},
		{
			name:          "failed-list",
			expectedError: "something went wrong",
			imageListFunc: func(options types.ImageListOptions) ([]types.ImageSummary, error) {
				return []types.ImageSummary{}, errors.Errorf("something went wrong")
			},
		},
	}
	for _, tc := range testCases {
		cmd := NewImagesCommand(test.NewFakeCli(&fakeClient{imageListFunc: tc.imageListFunc}))
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNewImagesCommandSuccess(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		imageFormat   string
		imageListFunc func(options types.ImageListOptions) ([]types.ImageSummary, error)
	}{
		{
			name: "simple",
		},
		{
			name:        "format",
			imageFormat: "raw",
		},
		{
			name:        "quiet-format",
			args:        []string{"-q"},
			imageFormat: "table",
		},
		{
			name: "match-name",
			args: []string{"image"},
			imageListFunc: func(options types.ImageListOptions) ([]types.ImageSummary, error) {
				assert.Check(t, is.Equal("image", options.Filters.Get("reference")[0]))
				return []types.ImageSummary{}, nil
			},
		},
		{
			name: "filters",
			args: []string{"--filter", "name=value"},
			imageListFunc: func(options types.ImageListOptions) ([]types.ImageSummary, error) {
				assert.Check(t, is.Equal("value", options.Filters.Get("name")[0]))
				return []types.ImageSummary{}, nil
			},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{imageListFunc: tc.imageListFunc})
		cli.SetConfigFile(&configfile.ConfigFile{ImagesFormat: tc.imageFormat})
		cmd := NewImagesCommand(cli)
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		assert.NilError(t, err)
		golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("list-command-success.%s.golden", tc.name))
	}
}

func TestNewListCommandAlias(t *testing.T) {
	cmd := newListCommand(test.NewFakeCli(&fakeClient{}))
	assert.Check(t, cmd.HasAlias("list"))
	assert.Check(t, !cmd.HasAlias("other"))
}
