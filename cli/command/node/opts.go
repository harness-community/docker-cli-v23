package node

import (
	"github.com/harness-community/docker-cli-v23/opts"
)

type nodeOptions struct {
	annotations
	role         string
	availability string
}

type annotations struct {
	labels opts.ListOpts
}

func newNodeOptions() *nodeOptions {
	return &nodeOptions{
		annotations: annotations{
			labels: opts.NewListOpts(nil),
		},
	}
}
