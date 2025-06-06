package image

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestNewLoadCommandErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		isTerminalIn  bool
		expectedError string
		imageLoadFunc func(input io.Reader, quiet bool) (types.ImageLoadResponse, error)
	}{
		{
			name:          "wrong-args",
			args:          []string{"arg"},
			expectedError: "accepts no arguments.",
		},
		{
			name:          "input-to-terminal",
			isTerminalIn:  true,
			expectedError: "requested load from stdin, but stdin is empty",
		},
		{
			name:          "pull-error",
			expectedError: "something went wrong",
			imageLoadFunc: func(input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
				return types.ImageLoadResponse{}, errors.Errorf("something went wrong")
			},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{imageLoadFunc: tc.imageLoadFunc})
		cli.In().SetIsTerminal(tc.isTerminalIn)
		cmd := NewLoadCommand(cli)
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNewLoadCommandInvalidInput(t *testing.T) {
	expectedError := "open *"
	cmd := NewLoadCommand(test.NewFakeCli(&fakeClient{}))
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"--input", "*"})
	err := cmd.Execute()
	assert.ErrorContains(t, err, expectedError)
}

func TestNewLoadCommandSuccess(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		imageLoadFunc func(input io.Reader, quiet bool) (types.ImageLoadResponse, error)
	}{
		{
			name: "simple",
			imageLoadFunc: func(input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
				return types.ImageLoadResponse{Body: io.NopCloser(strings.NewReader("Success"))}, nil
			},
		},
		{
			name: "json",
			imageLoadFunc: func(input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
				json := "{\"ID\": \"1\"}"
				return types.ImageLoadResponse{
					Body: io.NopCloser(strings.NewReader(json)),
					JSON: true,
				}, nil
			},
		},
		{
			name: "input-file",
			args: []string{"--input", "testdata/load-command-success.input.txt"},
			imageLoadFunc: func(input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
				return types.ImageLoadResponse{Body: io.NopCloser(strings.NewReader("Success"))}, nil
			},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{imageLoadFunc: tc.imageLoadFunc})
		cmd := NewLoadCommand(cli)
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		assert.NilError(t, err)
		golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("load-command-success.%s.golden", tc.name))
	}
}
