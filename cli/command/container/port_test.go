package container

import (
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/docker/go-connections/nat"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestNewPortCommandOutput(t *testing.T) {
	testCases := []struct {
		name string
		ips  []string
		port string
	}{
		{
			name: "container-port-ipv4",
			ips:  []string{"0.0.0.0"},
			port: "80",
		},
		{
			name: "container-port-ipv6",
			ips:  []string{"::"},
			port: "80",
		},
		{
			name: "container-port-ipv6-and-ipv4",
			ips:  []string{"::", "0.0.0.0"},
			port: "80",
		},
		{
			name: "container-port-ipv6-and-ipv4-443-udp",
			ips:  []string{"::", "0.0.0.0"},
			port: "443/udp",
		},
		{
			name: "container-port-all-ports",
			ips:  []string{"::", "0.0.0.0"},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cli := test.NewFakeCli(&fakeClient{
				inspectFunc: func(string) (types.ContainerJSON, error) {
					ci := types.ContainerJSON{NetworkSettings: &types.NetworkSettings{}}
					ci.NetworkSettings.Ports = nat.PortMap{
						"80/tcp":  make([]nat.PortBinding, len(tc.ips)),
						"443/tcp": make([]nat.PortBinding, len(tc.ips)),
						"443/udp": make([]nat.PortBinding, len(tc.ips)),
					}
					for i, ip := range tc.ips {
						ci.NetworkSettings.Ports["80/tcp"][i] = nat.PortBinding{
							HostIP: ip, HostPort: "3456",
						}
						ci.NetworkSettings.Ports["443/tcp"][i] = nat.PortBinding{
							HostIP: ip, HostPort: "4567",
						}
						ci.NetworkSettings.Ports["443/udp"][i] = nat.PortBinding{
							HostIP: ip, HostPort: "5678",
						}
					}
					return ci, nil
				},
			}, test.EnableContentTrust)
			cmd := NewPortCommand(cli)
			cmd.SetErr(io.Discard)
			cmd.SetArgs([]string{"some_container", tc.port})
			err := cmd.Execute()
			assert.NilError(t, err)
			golden.Assert(t, cli.OutBuffer().String(), tc.name+".golden")
		})
	}
}
