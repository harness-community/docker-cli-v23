package plugin

import (
	"fmt"
	"io"
	"runtime"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/fs"
)

func TestCreateErrors(t *testing.T) {
	noSuchFile := "no such file or directory"
	if runtime.GOOS == "windows" {
		noSuchFile = "The system cannot find the file specified."
	}
	testCases := []struct {
		args          []string
		expectedError string
	}{
		{
			args:          []string{},
			expectedError: "requires at least 2 arguments",
		},
		{
			args:          []string{"INVALID_TAG", "context-dir"},
			expectedError: "invalid",
		},
		{
			args:          []string{"plugin-foo", "nonexistent_context_dir"},
			expectedError: noSuchFile,
		},
	}

	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{})
		cmd := newCreateCommand(cli)
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestCreateErrorOnFileAsContextDir(t *testing.T) {
	tmpFile := fs.NewFile(t, "file-as-context-dir")
	defer tmpFile.Remove()

	cli := test.NewFakeCli(&fakeClient{})
	cmd := newCreateCommand(cli)
	cmd.SetArgs([]string{"plugin-foo", tmpFile.Path()})
	cmd.SetOut(io.Discard)
	assert.ErrorContains(t, cmd.Execute(), "context must be a directory")
}

func TestCreateErrorOnContextDirWithoutConfig(t *testing.T) {
	tmpDir := fs.NewDir(t, "plugin-create-test")
	defer tmpDir.Remove()

	cli := test.NewFakeCli(&fakeClient{})
	cmd := newCreateCommand(cli)
	cmd.SetArgs([]string{"plugin-foo", tmpDir.Path()})
	cmd.SetOut(io.Discard)

	expectedErr := "config.json: no such file or directory"
	if runtime.GOOS == "windows" {
		expectedErr = "config.json: The system cannot find the file specified."
	}
	assert.ErrorContains(t, cmd.Execute(), expectedErr)
}

func TestCreateErrorOnInvalidConfig(t *testing.T) {
	tmpDir := fs.NewDir(t, "plugin-create-test",
		fs.WithDir("rootfs"),
		fs.WithFile("config.json", "invalid-config-contents"))
	defer tmpDir.Remove()

	cli := test.NewFakeCli(&fakeClient{})
	cmd := newCreateCommand(cli)
	cmd.SetArgs([]string{"plugin-foo", tmpDir.Path()})
	cmd.SetOut(io.Discard)
	assert.ErrorContains(t, cmd.Execute(), "invalid")
}

func TestCreateErrorFromDaemon(t *testing.T) {
	tmpDir := fs.NewDir(t, "plugin-create-test",
		fs.WithDir("rootfs"),
		fs.WithFile("config.json", `{ "Name": "plugin-foo" }`))
	defer tmpDir.Remove()

	cli := test.NewFakeCli(&fakeClient{
		pluginCreateFunc: func(createContext io.Reader, createOptions types.PluginCreateOptions) error {
			return fmt.Errorf("Error creating plugin")
		},
	})

	cmd := newCreateCommand(cli)
	cmd.SetArgs([]string{"plugin-foo", tmpDir.Path()})
	cmd.SetOut(io.Discard)
	assert.ErrorContains(t, cmd.Execute(), "Error creating plugin")
}

func TestCreatePlugin(t *testing.T) {
	tmpDir := fs.NewDir(t, "plugin-create-test",
		fs.WithDir("rootfs"),
		fs.WithFile("config.json", `{ "Name": "plugin-foo" }`))
	defer tmpDir.Remove()

	cli := test.NewFakeCli(&fakeClient{
		pluginCreateFunc: func(createContext io.Reader, createOptions types.PluginCreateOptions) error {
			return nil
		},
	})

	cmd := newCreateCommand(cli)
	cmd.SetArgs([]string{"plugin-foo", tmpDir.Path()})
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal("plugin-foo\n", cli.OutBuffer().String()))
}
