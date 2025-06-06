package volume

import (
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
)

func TestVolumeRemoveErrors(t *testing.T) {
	testCases := []struct {
		args             []string
		volumeRemoveFunc func(volumeID string, force bool) error
		expectedError    string
	}{
		{
			expectedError: "requires at least 1 argument",
		},
		{
			args: []string{"nodeID"},
			volumeRemoveFunc: func(volumeID string, force bool) error {
				return errors.Errorf("error removing the volume")
			},
			expectedError: "error removing the volume",
		},
	}
	for _, tc := range testCases {
		cmd := newRemoveCommand(
			test.NewFakeCli(&fakeClient{
				volumeRemoveFunc: tc.volumeRemoveFunc,
			}))
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNodeRemoveMultiple(t *testing.T) {
	cmd := newRemoveCommand(test.NewFakeCli(&fakeClient{}))
	cmd.SetArgs([]string{"volume1", "volume2"})
	assert.NilError(t, cmd.Execute())
}
