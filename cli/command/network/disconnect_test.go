package network

import (
	"context"
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
)

func TestNetworkDisconnectErrors(t *testing.T) {
	testCases := []struct {
		args                  []string
		networkDisconnectFunc func(ctx context.Context, networkID, container string, force bool) error
		expectedError         string
	}{
		{
			expectedError: "requires exactly 2 arguments",
		},
		{
			args: []string{"toto", "titi"},
			networkDisconnectFunc: func(ctx context.Context, networkID, container string, force bool) error {
				return errors.Errorf("error disconnecting network")
			},
			expectedError: "error disconnecting network",
		},
	}

	for _, tc := range testCases {
		cmd := newDisconnectCommand(
			test.NewFakeCli(&fakeClient{
				networkDisconnectFunc: tc.networkDisconnectFunc,
			}),
		)
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}
