package stack

import (
	"io"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/config/configfile"
	"github.com/harness-community/docker-cli-v23/internal/test"
	. "github.com/harness-community/docker-cli-v23/internal/test/builders" // Import builders to get the builder function as package function
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestStackServicesErrors(t *testing.T) {
	testCases := []struct {
		args            []string
		flags           map[string]string
		serviceListFunc func(options types.ServiceListOptions) ([]swarm.Service, error)
		nodeListFunc    func(options types.NodeListOptions) ([]swarm.Node, error)
		taskListFunc    func(options types.TaskListOptions) ([]swarm.Task, error)
		expectedError   string
	}{
		{
			args: []string{"foo"},
			serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
				return nil, errors.Errorf("error getting services")
			},
			expectedError: "error getting services",
		},
		{
			args: []string{"foo"},
			serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
				return []swarm.Service{*Service(GlobalService())}, nil
			},
			nodeListFunc: func(options types.NodeListOptions) ([]swarm.Node, error) {
				return nil, errors.Errorf("error getting nodes")
			},
			taskListFunc: func(options types.TaskListOptions) ([]swarm.Task, error) {
				return []swarm.Task{*Task()}, nil
			},
			expectedError: "error getting nodes",
		},
		{
			args: []string{"foo"},
			serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
				return []swarm.Service{*Service(GlobalService())}, nil
			},
			taskListFunc: func(options types.TaskListOptions) ([]swarm.Task, error) {
				return nil, errors.Errorf("error getting tasks")
			},
			expectedError: "error getting tasks",
		},
		{
			args: []string{"foo"},
			flags: map[string]string{
				"format": "{{invalid format}}",
			},
			serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
				return []swarm.Service{*Service()}, nil
			},
			expectedError: "template parsing error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.expectedError, func(t *testing.T) {
			cli := test.NewFakeCli(&fakeClient{
				serviceListFunc: tc.serviceListFunc,
				nodeListFunc:    tc.nodeListFunc,
				taskListFunc:    tc.taskListFunc,
			})
			cmd := newServicesCommand(cli)
			cmd.SetArgs(tc.args)
			for key, value := range tc.flags {
				cmd.Flags().Set(key, value)
			}
			cmd.SetOut(io.Discard)
			assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
		})
	}
}

func TestRunServicesWithEmptyName(t *testing.T) {
	cmd := newServicesCommand(test.NewFakeCli(&fakeClient{}))
	cmd.SetArgs([]string{"'   '"})
	cmd.SetOut(io.Discard)

	assert.ErrorContains(t, cmd.Execute(), `invalid stack name: "'   '"`)
}

func TestStackServicesEmptyServiceList(t *testing.T) {
	fakeCli := test.NewFakeCli(&fakeClient{
		serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
			return []swarm.Service{}, nil
		},
	})
	cmd := newServicesCommand(fakeCli)
	cmd.SetArgs([]string{"foo"})
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal("", fakeCli.OutBuffer().String()))
	assert.Check(t, is.Equal("Nothing found in stack: foo\n", fakeCli.ErrBuffer().String()))
}

func TestStackServicesWithQuietOption(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
			return []swarm.Service{*Service(ServiceID("id-foo"))}, nil
		},
	})
	cmd := newServicesCommand(cli)
	cmd.Flags().Set("quiet", "true")
	cmd.SetArgs([]string{"foo"})
	assert.NilError(t, cmd.Execute())
	golden.Assert(t, cli.OutBuffer().String(), "stack-services-with-quiet-option.golden")
}

func TestStackServicesWithFormat(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
			return []swarm.Service{
				*Service(ServiceName("service-name-foo")),
			}, nil
		},
	})
	cmd := newServicesCommand(cli)
	cmd.SetArgs([]string{"foo"})
	cmd.Flags().Set("format", "{{ .Name }}")
	assert.NilError(t, cmd.Execute())
	golden.Assert(t, cli.OutBuffer().String(), "stack-services-with-format.golden")
}

func TestStackServicesWithConfigFormat(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
			return []swarm.Service{
				*Service(ServiceName("service-name-foo")),
			}, nil
		},
	})
	cli.SetConfigFile(&configfile.ConfigFile{
		ServicesFormat: "{{ .Name }}",
	})
	cmd := newServicesCommand(cli)
	cmd.SetArgs([]string{"foo"})
	assert.NilError(t, cmd.Execute())
	golden.Assert(t, cli.OutBuffer().String(), "stack-services-with-config-format.golden")
}

func TestStackServicesWithoutFormat(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		serviceListFunc: func(options types.ServiceListOptions) ([]swarm.Service, error) {
			return []swarm.Service{*Service(
				ServiceName("name-foo"),
				ServiceID("id-foo"),
				ReplicatedService(2),
				ServiceImage("busybox:latest"),
				ServicePort(swarm.PortConfig{
					PublishMode:   swarm.PortConfigPublishModeIngress,
					PublishedPort: 0,
					TargetPort:    3232,
					Protocol:      swarm.PortConfigProtocolTCP,
				}),
			)}, nil
		},
	})
	cmd := newServicesCommand(cli)
	cmd.SetArgs([]string{"foo"})
	assert.NilError(t, cmd.Execute())
	golden.Assert(t, cli.OutBuffer().String(), "stack-services-without-format.golden")
}
