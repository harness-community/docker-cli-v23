package volume // import "github.com/harness-community/docker-v23/api/types/volume"

import "github.com/harness-community/docker-v23/api/types/filters"

// ListOptions holds parameters to list volumes.
type ListOptions struct {
	Filters filters.Args
}
