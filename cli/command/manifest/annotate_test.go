package manifest

import (
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/manifest/store"
	"github.com/harness-community/docker-cli-v23/internal/test"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestManifestAnnotateError(t *testing.T) {
	testCases := []struct {
		args          []string
		expectedError string
	}{
		{
			args:          []string{"too-few-arguments"},
			expectedError: "requires exactly 2 arguments",
		},
		{
			args:          []string{"th!si'sa/fa!ke/li$t/name", "example.com/alpine:3.0"},
			expectedError: "error parsing name for manifest list",
		},
		{
			args:          []string{"example.com/list:v1", "th!si'sa/fa!ke/im@ge/nam32"},
			expectedError: "error parsing name for manifest",
		},
	}

	for _, tc := range testCases {
		cli := test.NewFakeCli(nil)
		cmd := newAnnotateCommand(cli)
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestManifestAnnotate(t *testing.T) {
	store := store.NewStore(t.TempDir())

	cli := test.NewFakeCli(nil)
	cli.SetManifestStore(store)
	namedRef := ref(t, "alpine:3.0")
	imageManifest := fullImageManifest(t, namedRef)
	err := store.Save(ref(t, "list:v1"), namedRef, imageManifest)
	assert.NilError(t, err)

	cmd := newAnnotateCommand(cli)
	cmd.SetArgs([]string{"example.com/list:v1", "example.com/fake:0.0"})
	cmd.SetOut(io.Discard)
	expectedError := "manifest for image example.com/fake:0.0 does not exist"
	assert.ErrorContains(t, cmd.Execute(), expectedError)

	cmd.SetArgs([]string{"example.com/list:v1", "example.com/alpine:3.0"})
	cmd.Flags().Set("os", "freebsd")
	cmd.Flags().Set("arch", "fake")
	cmd.Flags().Set("os-version", "1")
	cmd.Flags().Set("os-features", "feature1")
	cmd.Flags().Set("variant", "v7")
	expectedError = "manifest entry for image has unsupported os/arch combination"
	assert.ErrorContains(t, cmd.Execute(), expectedError)

	cmd.Flags().Set("arch", "arm")
	assert.NilError(t, cmd.Execute())

	cmd = newInspectCommand(cli)
	err = cmd.Flags().Set("verbose", "true")
	assert.NilError(t, err)
	cmd.SetArgs([]string{"example.com/list:v1", "example.com/alpine:3.0"})
	assert.NilError(t, cmd.Execute())
	actual := cli.OutBuffer()
	expected := golden.Get(t, "inspect-annotate.golden")
	assert.Check(t, is.Equal(string(expected), actual.String()))
}
