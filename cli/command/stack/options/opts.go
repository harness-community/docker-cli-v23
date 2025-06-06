package options

import "github.com/harness-community/docker-cli-v23/opts"

// Deploy holds docker stack deploy options
type Deploy struct {
	Composefiles     []string
	Namespace        string
	ResolveImage     string
	SendRegistryAuth bool
	Prune            bool
}

// Config holds docker stack config options
type Config struct {
	Composefiles      []string
	SkipInterpolation bool
}

// List holds docker stack ls options
type List struct {
	Format        string
	AllNamespaces bool
}

// PS holds docker stack ps options
type PS struct {
	Filter    opts.FilterOpt
	NoTrunc   bool
	Namespace string
	NoResolve bool
	Quiet     bool
	Format    string
}

// Remove holds docker stack remove options
type Remove struct {
	Namespaces []string
}

// Services holds docker stack services options
type Services struct {
	Quiet     bool
	Format    string
	Filter    opts.FilterOpt
	Namespace string
}
