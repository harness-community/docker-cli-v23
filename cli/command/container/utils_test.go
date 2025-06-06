package container

import (
	"context"
	"strings"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api"
	"github.com/harness-community/docker-v23/api/types/container"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func waitFn(cid string) (<-chan container.WaitResponse, <-chan error) {
	resC := make(chan container.WaitResponse)
	errC := make(chan error, 1)
	var res container.WaitResponse

	go func() {
		switch {
		case strings.Contains(cid, "exit-code-42"):
			res.StatusCode = 42
			resC <- res
		case strings.Contains(cid, "non-existent"):
			err := errors.Errorf("no such container: %v", cid)
			errC <- err
		case strings.Contains(cid, "wait-error"):
			res.Error = &container.WaitExitError{Message: "removal failed"}
			resC <- res
		default:
			// normal exit
			resC <- res
		}
	}()

	return resC, errC
}

func TestWaitExitOrRemoved(t *testing.T) {
	testcases := []struct {
		cid      string
		exitCode int
	}{
		{
			cid:      "normal-container",
			exitCode: 0,
		},
		{
			cid:      "give-me-exit-code-42",
			exitCode: 42,
		},
		{
			cid:      "i-want-a-wait-error",
			exitCode: 125,
		},
		{
			cid:      "non-existent-container-id",
			exitCode: 125,
		},
	}

	client := test.NewFakeCli(&fakeClient{waitFunc: waitFn, Version: api.DefaultVersion})
	for _, testcase := range testcases {
		statusC := waitExitOrRemoved(context.Background(), client, testcase.cid, true)
		exitCode := <-statusC
		assert.Check(t, is.Equal(testcase.exitCode, exitCode))
	}
}
