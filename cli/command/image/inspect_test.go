package image

import (
	"fmt"
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestNewInspectCommandErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "wrong-args",
			args:          []string{},
			expectedError: "requires at least 1 argument.",
		},
	}
	for _, tc := range testCases {
		cmd := newInspectCommand(test.NewFakeCli(&fakeClient{}))
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNewInspectCommandSuccess(t *testing.T) {
	imageInspectInvocationCount := 0
	testCases := []struct {
		name             string
		args             []string
		imageCount       int
		imageInspectFunc func(image string) (types.ImageInspect, []byte, error)
	}{
		{
			name:       "simple",
			args:       []string{"image"},
			imageCount: 1,
			imageInspectFunc: func(image string) (types.ImageInspect, []byte, error) {
				imageInspectInvocationCount++
				assert.Check(t, is.Equal("image", image))
				return types.ImageInspect{}, nil, nil
			},
		},
		{
			name:       "format",
			imageCount: 1,
			args:       []string{"--format='{{.ID}}'", "image"},
			imageInspectFunc: func(image string) (types.ImageInspect, []byte, error) {
				imageInspectInvocationCount++
				return types.ImageInspect{ID: image}, nil, nil
			},
		},
		{
			name:       "simple-many",
			args:       []string{"image1", "image2"},
			imageCount: 2,
			imageInspectFunc: func(image string) (types.ImageInspect, []byte, error) {
				imageInspectInvocationCount++
				if imageInspectInvocationCount == 1 {
					assert.Check(t, is.Equal("image1", image))
				} else {
					assert.Check(t, is.Equal("image2", image))
				}
				return types.ImageInspect{}, nil, nil
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			imageInspectInvocationCount = 0
			cli := test.NewFakeCli(&fakeClient{imageInspectFunc: tc.imageInspectFunc})
			cmd := newInspectCommand(cli)
			cmd.SetOut(io.Discard)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			assert.NilError(t, err)
			golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("inspect-command-success.%s.golden", tc.name))
			assert.Check(t, is.Equal(imageInspectInvocationCount, tc.imageCount))
		})
	}
}
