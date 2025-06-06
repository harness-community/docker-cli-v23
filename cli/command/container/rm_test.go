package container

import (
	"context"
	"fmt"
	"io"
	"sort"
	"sync"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/errdefs"
	"gotest.tools/v3/assert"
)

func TestRemoveForce(t *testing.T) {
	for _, tc := range []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{name: "without force", args: []string{"nosuchcontainer", "mycontainer"}, expectedErr: "no such container"},
		{name: "with force", args: []string{"--force", "nosuchcontainer", "mycontainer"}, expectedErr: ""},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var removed []string
			mutex := new(sync.Mutex)

			cli := test.NewFakeCli(&fakeClient{
				containerRemoveFunc: func(ctx context.Context, container string, options types.ContainerRemoveOptions) error {
					// containerRemoveFunc is called in parallel for each container
					// by the remove command so append must be synchronized.
					mutex.Lock()
					removed = append(removed, container)
					mutex.Unlock()

					if container == "nosuchcontainer" {
						return errdefs.NotFound(fmt.Errorf("Error: no such container: " + container))
					}
					return nil
				},
				Version: "1.36",
			})
			cmd := NewRmCommand(cli)
			cmd.SetOut(io.Discard)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NilError(t, err)
			}
			sort.Strings(removed)
			assert.DeepEqual(t, removed, []string{"mycontainer", "nosuchcontainer"})
		})
	}
}
