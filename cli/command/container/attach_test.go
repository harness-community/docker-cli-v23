package container

import (
	"fmt"
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli"
	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/container"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
)

func TestNewAttachCommandErrors(t *testing.T) {
	testCases := []struct {
		name                 string
		args                 []string
		expectedError        string
		containerInspectFunc func(img string) (types.ContainerJSON, error)
	}{
		{
			name:          "client-error",
			args:          []string{"5cb5bb5e4a3b"},
			expectedError: "something went wrong",
			containerInspectFunc: func(containerID string) (types.ContainerJSON, error) {
				return types.ContainerJSON{}, errors.Errorf("something went wrong")
			},
		},
		{
			name:          "client-stopped",
			args:          []string{"5cb5bb5e4a3b"},
			expectedError: "You cannot attach to a stopped container",
			containerInspectFunc: func(containerID string) (types.ContainerJSON, error) {
				c := types.ContainerJSON{}
				c.ContainerJSONBase = &types.ContainerJSONBase{}
				c.ContainerJSONBase.State = &types.ContainerState{Running: false}
				return c, nil
			},
		},
		{
			name:          "client-paused",
			args:          []string{"5cb5bb5e4a3b"},
			expectedError: "You cannot attach to a paused container",
			containerInspectFunc: func(containerID string) (types.ContainerJSON, error) {
				c := types.ContainerJSON{}
				c.ContainerJSONBase = &types.ContainerJSONBase{}
				c.ContainerJSONBase.State = &types.ContainerState{
					Running: true,
					Paused:  true,
				}
				return c, nil
			},
		},
		{
			name:          "client-restarting",
			args:          []string{"5cb5bb5e4a3b"},
			expectedError: "You cannot attach to a restarting container",
			containerInspectFunc: func(containerID string) (types.ContainerJSON, error) {
				c := types.ContainerJSON{}
				c.ContainerJSONBase = &types.ContainerJSONBase{}
				c.ContainerJSONBase.State = &types.ContainerState{
					Running:    true,
					Paused:     false,
					Restarting: true,
				}
				return c, nil
			},
		},
	}
	for _, tc := range testCases {
		cmd := NewAttachCommand(test.NewFakeCli(&fakeClient{inspectFunc: tc.containerInspectFunc}))
		cmd.SetOut(io.Discard)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestGetExitStatus(t *testing.T) {
	var (
		expectedErr = fmt.Errorf("unexpected error")
		errC        = make(chan error, 1)
		resultC     = make(chan container.WaitResponse, 1)
	)

	testcases := []struct {
		result        *container.WaitResponse
		err           error
		expectedError error
	}{
		{
			result: &container.WaitResponse{
				StatusCode: 0,
			},
		},
		{
			err:           expectedErr,
			expectedError: expectedErr,
		},
		{
			result: &container.WaitResponse{
				Error: &container.WaitExitError{Message: expectedErr.Error()},
			},
			expectedError: expectedErr,
		},
		{
			result: &container.WaitResponse{
				StatusCode: 15,
			},
			expectedError: cli.StatusError{StatusCode: 15},
		},
	}

	for _, testcase := range testcases {
		if testcase.err != nil {
			errC <- testcase.err
		}
		if testcase.result != nil {
			resultC <- *testcase.result
		}
		err := getExitStatus(errC, resultC)
		if testcase.expectedError == nil {
			assert.NilError(t, err)
		} else {
			assert.Error(t, err, testcase.expectedError.Error())
		}
	}
}
