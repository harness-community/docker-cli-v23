package container

import (
	"context"
	"testing"
	"time"

	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestInitTtySizeErrors(t *testing.T) {
	expectedError := "failed to resize tty, using default size\n"
	fakeContainerExecResizeFunc := func(id string, options types.ResizeOptions) error {
		return errors.Errorf("Error response from daemon: no such exec")
	}
	fakeResizeTtyFunc := func(ctx context.Context, cli command.Cli, id string, isExec bool) error {
		height, width := uint(1024), uint(768)
		return resizeTtyTo(ctx, cli.Client(), id, height, width, isExec)
	}
	ctx := context.Background()
	cli := test.NewFakeCli(&fakeClient{containerExecResizeFunc: fakeContainerExecResizeFunc})
	initTtySize(ctx, cli, "8mm8nn8tt8bb", true, fakeResizeTtyFunc)
	time.Sleep(1500 * time.Millisecond)
	assert.Check(t, is.Equal(expectedError, cli.ErrBuffer().String()))
}
