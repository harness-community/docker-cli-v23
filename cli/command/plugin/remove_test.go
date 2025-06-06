package plugin

import (
	"fmt"
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestRemoveErrors(t *testing.T) {
	testCases := []struct {
		args             []string
		pluginRemoveFunc func(name string, options types.PluginRemoveOptions) error
		expectedError    string
	}{
		{
			args:          []string{},
			expectedError: "requires at least 1 argument",
		},
		{
			args: []string{"plugin-foo"},
			pluginRemoveFunc: func(name string, options types.PluginRemoveOptions) error {
				return fmt.Errorf("Error removing plugin")
			},
			expectedError: "Error removing plugin",
		},
	}

	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			pluginRemoveFunc: tc.pluginRemoveFunc,
		})
		cmd := newRemoveCommand(cli)
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestRemove(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		pluginRemoveFunc: func(name string, options types.PluginRemoveOptions) error {
			return nil
		},
	})
	cmd := newRemoveCommand(cli)
	cmd.SetArgs([]string{"plugin-foo"})
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal("plugin-foo\n", cli.OutBuffer().String()))
}

func TestRemoveWithForceOption(t *testing.T) {
	force := false
	cli := test.NewFakeCli(&fakeClient{
		pluginRemoveFunc: func(name string, options types.PluginRemoveOptions) error {
			force = options.Force
			return nil
		},
	})
	cmd := newRemoveCommand(cli)
	cmd.SetArgs([]string{"plugin-foo"})
	cmd.Flags().Set("force", "true")
	assert.NilError(t, cmd.Execute())
	assert.Check(t, force)
	assert.Check(t, is.Equal("plugin-foo\n", cli.OutBuffer().String()))
}
