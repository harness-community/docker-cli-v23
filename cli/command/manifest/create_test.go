package manifest

import (
	"context"
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/manifest/store"
	manifesttypes "github.com/harness-community/docker-cli-v23/cli/manifest/types"
	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestManifestCreateErrors(t *testing.T) {
	testCases := []struct {
		args          []string
		expectedError string
	}{
		{
			args:          []string{"too-few-arguments"},
			expectedError: "requires at least 2 arguments",
		},
		{
			args:          []string{"th!si'sa/fa!ke/li$t/name", "example.com/alpine:3.0"},
			expectedError: "error parsing name for manifest list",
		},
	}

	for _, tc := range testCases {
		cli := test.NewFakeCli(nil)
		cmd := newCreateListCommand(cli)
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

// create a manifest list, then overwrite it, and inspect to see if the old one is still there
func TestManifestCreateAmend(t *testing.T) {
	store := store.NewStore(t.TempDir())

	cli := test.NewFakeCli(nil)
	cli.SetManifestStore(store)

	namedRef := ref(t, "alpine:3.0")
	imageManifest := fullImageManifest(t, namedRef)
	err := store.Save(ref(t, "list:v1"), namedRef, imageManifest)
	assert.NilError(t, err)
	namedRef = ref(t, "alpine:3.1")
	imageManifest = fullImageManifest(t, namedRef)
	err = store.Save(ref(t, "list:v1"), namedRef, imageManifest)
	assert.NilError(t, err)

	cmd := newCreateListCommand(cli)
	cmd.SetArgs([]string{"example.com/list:v1", "example.com/alpine:3.1"})
	cmd.Flags().Set("amend", "true")
	cmd.SetOut(io.Discard)
	err = cmd.Execute()
	assert.NilError(t, err)

	// make a new cli to clear the buffers
	cli = test.NewFakeCli(nil)
	cli.SetManifestStore(store)
	inspectCmd := newInspectCommand(cli)
	inspectCmd.SetArgs([]string{"example.com/list:v1"})
	assert.NilError(t, inspectCmd.Execute())
	actual := cli.OutBuffer()
	expected := golden.Get(t, "inspect-manifest-list.golden")
	assert.Check(t, is.Equal(string(expected), actual.String()))
}

// attempt to overwrite a saved manifest and get refused
func TestManifestCreateRefuseAmend(t *testing.T) {
	store := store.NewStore(t.TempDir())

	cli := test.NewFakeCli(nil)
	cli.SetManifestStore(store)
	namedRef := ref(t, "alpine:3.0")
	imageManifest := fullImageManifest(t, namedRef)
	err := store.Save(ref(t, "list:v1"), namedRef, imageManifest)
	assert.NilError(t, err)

	cmd := newCreateListCommand(cli)
	cmd.SetArgs([]string{"example.com/list:v1", "example.com/alpine:3.0"})
	cmd.SetOut(io.Discard)
	err = cmd.Execute()
	assert.Error(t, err, "refusing to amend an existing manifest list with no --amend flag")
}

// attempt to make a manifest list without valid images
func TestManifestCreateNoManifest(t *testing.T) {
	store := store.NewStore(t.TempDir())

	cli := test.NewFakeCli(nil)
	cli.SetManifestStore(store)
	cli.SetRegistryClient(&fakeRegistryClient{
		getManifestFunc: func(_ context.Context, ref reference.Named) (manifesttypes.ImageManifest, error) {
			return manifesttypes.ImageManifest{}, errors.Errorf("No such image: %v", ref)
		},
		getManifestListFunc: func(ctx context.Context, ref reference.Named) ([]manifesttypes.ImageManifest, error) {
			return nil, errors.Errorf("No such manifest: %s", ref)
		},
	})

	cmd := newCreateListCommand(cli)
	cmd.SetArgs([]string{"example.com/list:v1", "example.com/alpine:3.0"})
	cmd.SetOut(io.Discard)
	err := cmd.Execute()
	assert.Error(t, err, "No such image: example.com/alpine:3.0")
}
