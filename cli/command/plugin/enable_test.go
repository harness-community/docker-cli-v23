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

func TestPluginEnableErrors(t *testing.T) {
	testCases := []struct {
		args             []string
		flags            map[string]string
		pluginEnableFunc func(name string, options types.PluginEnableOptions) error
		expectedError    string
	}{
		{
			args:          []string{},
			expectedError: "requires exactly 1 argument",
		},
		{
			args:          []string{"too-many", "arguments"},
			expectedError: "requires exactly 1 argument",
		},
		{
			args: []string{"plugin-foo"},
			pluginEnableFunc: func(name string, options types.PluginEnableOptions) error {
				return fmt.Errorf("failed to enable plugin")
			},
			expectedError: "failed to enable plugin",
		},
		{
			args: []string{"plugin-foo"},
			flags: map[string]string{
				"timeout": "-1",
			},
			expectedError: "negative timeout -1 is invalid",
		},
	}

	for _, tc := range testCases {
		cmd := newEnableCommand(
			test.NewFakeCli(&fakeClient{
				pluginEnableFunc: tc.pluginEnableFunc,
			}))
		cmd.SetArgs(tc.args)
		for key, value := range tc.flags {
			cmd.Flags().Set(key, value)
		}
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestPluginEnable(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		pluginEnableFunc: func(name string, options types.PluginEnableOptions) error {
			return nil
		},
	})

	cmd := newEnableCommand(cli)
	cmd.SetArgs([]string{"plugin-foo"})
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal("plugin-foo\n", cli.OutBuffer().String()))
}
