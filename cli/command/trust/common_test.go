package trust

import (
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/trust"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/tuf/data"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMatchReleasedSignaturesSortOrder(t *testing.T) {
	releasesRole := data.DelegationRole{BaseRole: data.BaseRole{Name: trust.ReleasesRole}}
	targets := []client.TargetSignedStruct{
		{Target: client.Target{Name: "target10-foo"}, Role: releasesRole},
		{Target: client.Target{Name: "target1-foo"}, Role: releasesRole},
		{Target: client.Target{Name: "target2-foo"}, Role: releasesRole},
	}

	rows := matchReleasedSignatures(targets)

	var targetNames []string
	for _, r := range rows {
		targetNames = append(targetNames, r.SignedTag)
	}
	expected := []string{
		"target1-foo",
		"target2-foo",
		"target10-foo",
	}
	assert.Check(t, is.DeepEqual(expected, targetNames))
}
