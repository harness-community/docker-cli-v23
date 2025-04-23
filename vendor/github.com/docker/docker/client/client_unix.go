//go:build linux || freebsd || openbsd || netbsd || darwin || solaris || illumos || dragonfly
// +build linux freebsd openbsd netbsd darwin solaris illumos dragonfly

package client // import "github.com/harness-community/docker-v23/client"

// DefaultDockerHost defines OS-specific default host if the DOCKER_HOST
// (EnvOverrideHost) environment variable is unset or empty.
const DefaultDockerHost = "unix:///var/run/docker.sock"

const defaultProto = "unix"
const defaultAddr = "/var/run/docker.sock"
