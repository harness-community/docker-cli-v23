package builders

import (
	"time"

	"github.com/harness-community/docker-v23/api/types"
)

// Container creates a container with default values.
// Any number of container function builder can be passed to augment it.
func Container(name string, builders ...func(container *types.Container)) *types.Container {
	// now := time.Now()
	// onehourago := now.Add(-120 * time.Minute)
	container := &types.Container{
		ID:      "container_id",
		Names:   []string{"/" + name},
		Command: "top",
		Image:   "busybox:latest",
		Status:  "Up 1 minute",
		Created: time.Now().Add(-1 * time.Minute).Unix(),
	}

	for _, builder := range builders {
		builder(container)
	}

	return container
}

// WithLabel adds a label to the container
func WithLabel(key, value string) func(*types.Container) {
	return func(c *types.Container) {
		if c.Labels == nil {
			c.Labels = map[string]string{}
		}
		c.Labels[key] = value
	}
}

// WithName adds a name to the container
func WithName(name string) func(*types.Container) {
	return func(c *types.Container) {
		c.Names = append(c.Names, "/"+name)
	}
}

// WithPort adds a port mapping to the container
func WithPort(privateport, publicport uint16, builders ...func(*types.Port)) func(*types.Container) {
	return func(c *types.Container) {
		if c.Ports == nil {
			c.Ports = []types.Port{}
		}
		port := &types.Port{
			PrivatePort: privateport,
			PublicPort:  publicport,
		}
		for _, builder := range builders {
			builder(port)
		}
		c.Ports = append(c.Ports, *port)
	}
}

// WithSize adds size in bytes to the container
func WithSize(size int64) func(*types.Container) {
	return func(c *types.Container) {
		if size >= 0 {
			c.SizeRw = size
		}
	}
}

// IP sets the ip of the port
func IP(ip string) func(*types.Port) {
	return func(p *types.Port) {
		p.IP = ip
	}
}

// TCP sets the port to tcp
func TCP(p *types.Port) {
	p.Type = "tcp"
}

// UDP sets the port to udp
func UDP(p *types.Port) {
	p.Type = "udp"
}
