package stack

import (
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"gotest.tools/v3/assert"
)

func TestDeployWithEmptyName(t *testing.T) {
	cmd := newDeployCommand(test.NewFakeCli(&fakeClient{}))
	cmd.SetArgs([]string{"'   '"})
	cmd.SetOut(io.Discard)

	assert.ErrorContains(t, cmd.Execute(), `invalid stack name: "'   '"`)
}
