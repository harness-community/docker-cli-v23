package checkpoint

import (
	"bytes"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	"github.com/harness-community/docker-v23/api/types"
	"gotest.tools/v3/assert"
)

func TestCheckpointContextFormatWrite(t *testing.T) {
	cases := []struct {
		context  formatter.Context
		expected string
	}{
		{
			formatter.Context{Format: NewFormat(defaultCheckpointFormat)},
			`CHECKPOINT NAME
checkpoint-1
checkpoint-2
checkpoint-3
`,
		},
		{
			formatter.Context{Format: NewFormat("{{.Name}}")},
			`checkpoint-1
checkpoint-2
checkpoint-3
`,
		},
		{
			formatter.Context{Format: NewFormat("{{.Name}}:")},
			`checkpoint-1:
checkpoint-2:
checkpoint-3:
`,
		},
	}

	checkpoints := []types.Checkpoint{
		{Name: "checkpoint-1"},
		{Name: "checkpoint-2"},
		{Name: "checkpoint-3"},
	}
	for _, testcase := range cases {
		out := bytes.NewBufferString("")
		testcase.context.Output = out
		err := FormatWrite(testcase.context, checkpoints)
		assert.NilError(t, err)
		assert.Equal(t, out.String(), testcase.expected)
	}
}
