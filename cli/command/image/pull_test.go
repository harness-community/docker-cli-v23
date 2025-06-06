package image

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-cli-v23/internal/test/notary"
	"github.com/harness-community/docker-v23/api/types"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestNewPullCommandErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "wrong-args",
			expectedError: "requires exactly 1 argument.",
			args:          []string{},
		},
		{
			name:          "invalid-name",
			expectedError: "invalid reference format: repository name must be lowercase",
			args:          []string{"UPPERCASE_REPO"},
		},
		{
			name:          "all-tags-with-tag",
			expectedError: "tag can't be used with --all-tags/-a",
			args:          []string{"--all-tags", "image:tag"},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{})
		cmd := NewPullCommand(cli)
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNewPullCommandSuccess(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectedTag string
	}{
		{
			name:        "simple",
			args:        []string{"image:tag"},
			expectedTag: "image:tag",
		},
		{
			name:        "simple-no-tag",
			args:        []string{"image"},
			expectedTag: "image:latest",
		},
		{
			name:        "simple-quiet",
			args:        []string{"--quiet", "image"},
			expectedTag: "image:latest",
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			imagePullFunc: func(ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
				assert.Check(t, is.Equal(tc.expectedTag, ref), tc.name)
				return io.NopCloser(strings.NewReader("")), nil
			},
		})
		cmd := NewPullCommand(cli)
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		assert.NilError(t, err)
		golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("pull-command-success.%s.golden", tc.name))
	}
}

func TestNewPullCommandWithContentTrustErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedError string
		notaryFunc    test.NotaryClientFuncType
	}{
		{
			name:          "offline-notary-server",
			notaryFunc:    notary.GetOfflineNotaryRepository,
			expectedError: "client is offline",
			args:          []string{"image:tag"},
		},
		{
			name:          "uninitialized-notary-server",
			notaryFunc:    notary.GetUninitializedNotaryRepository,
			expectedError: "remote trust data does not exist",
			args:          []string{"image:tag"},
		},
		{
			name:          "empty-notary-server",
			notaryFunc:    notary.GetEmptyTargetsNotaryRepository,
			expectedError: "No valid trust data for tag",
			args:          []string{"image:tag"},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			imagePullFunc: func(ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("")), fmt.Errorf("shouldn't try to pull image")
			},
		}, test.EnableContentTrust)
		cli.SetNotaryClient(tc.notaryFunc)
		cmd := NewPullCommand(cli)
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		assert.ErrorContains(t, err, tc.expectedError)
	}
}
