package container

import (
	"context"
	"io"

	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/container"
	"github.com/harness-community/docker-v23/api/types/network"
	"github.com/harness-community/docker-v23/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type fakeClient struct {
	client.Client
	inspectFunc         func(string) (types.ContainerJSON, error)
	execInspectFunc     func(execID string) (types.ContainerExecInspect, error)
	execCreateFunc      func(container string, config types.ExecConfig) (types.IDResponse, error)
	createContainerFunc func(config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *specs.Platform,
		containerName string) (container.CreateResponse, error)
	containerStartFunc      func(container string, options types.ContainerStartOptions) error
	imageCreateFunc         func(parentReference string, options types.ImageCreateOptions) (io.ReadCloser, error)
	infoFunc                func() (types.Info, error)
	containerStatPathFunc   func(container, path string) (types.ContainerPathStat, error)
	containerCopyFromFunc   func(container, srcPath string) (io.ReadCloser, types.ContainerPathStat, error)
	logFunc                 func(string, types.ContainerLogsOptions) (io.ReadCloser, error)
	waitFunc                func(string) (<-chan container.WaitResponse, <-chan error)
	containerListFunc       func(types.ContainerListOptions) ([]types.Container, error)
	containerExportFunc     func(string) (io.ReadCloser, error)
	containerExecResizeFunc func(id string, options types.ResizeOptions) error
	containerRemoveFunc     func(ctx context.Context, container string, options types.ContainerRemoveOptions) error
	containerKillFunc       func(ctx context.Context, container, signal string) error
	Version                 string
}

func (f *fakeClient) ContainerList(_ context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	if f.containerListFunc != nil {
		return f.containerListFunc(options)
	}
	return []types.Container{}, nil
}

func (f *fakeClient) ContainerInspect(_ context.Context, containerID string) (types.ContainerJSON, error) {
	if f.inspectFunc != nil {
		return f.inspectFunc(containerID)
	}
	return types.ContainerJSON{}, nil
}

func (f *fakeClient) ContainerExecCreate(_ context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	if f.execCreateFunc != nil {
		return f.execCreateFunc(container, config)
	}
	return types.IDResponse{}, nil
}

func (f *fakeClient) ContainerExecInspect(_ context.Context, execID string) (types.ContainerExecInspect, error) {
	if f.execInspectFunc != nil {
		return f.execInspectFunc(execID)
	}
	return types.ContainerExecInspect{}, nil
}

func (f *fakeClient) ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error {
	return nil
}

func (f *fakeClient) ContainerCreate(
	_ context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	platform *specs.Platform,
	containerName string,
) (container.CreateResponse, error) {
	if f.createContainerFunc != nil {
		return f.createContainerFunc(config, hostConfig, networkingConfig, platform, containerName)
	}
	return container.CreateResponse{}, nil
}

func (f *fakeClient) ContainerRemove(ctx context.Context, container string, options types.ContainerRemoveOptions) error {
	if f.containerRemoveFunc != nil {
		return f.containerRemoveFunc(ctx, container, options)
	}
	return nil
}

func (f *fakeClient) ImageCreate(ctx context.Context, parentReference string, options types.ImageCreateOptions) (io.ReadCloser, error) {
	if f.imageCreateFunc != nil {
		return f.imageCreateFunc(parentReference, options)
	}
	return nil, nil
}

func (f *fakeClient) Info(_ context.Context) (types.Info, error) {
	if f.infoFunc != nil {
		return f.infoFunc()
	}
	return types.Info{}, nil
}

func (f *fakeClient) ContainerStatPath(_ context.Context, container, path string) (types.ContainerPathStat, error) {
	if f.containerStatPathFunc != nil {
		return f.containerStatPathFunc(container, path)
	}
	return types.ContainerPathStat{}, nil
}

func (f *fakeClient) CopyFromContainer(_ context.Context, container, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	if f.containerCopyFromFunc != nil {
		return f.containerCopyFromFunc(container, srcPath)
	}
	return nil, types.ContainerPathStat{}, nil
}

func (f *fakeClient) ContainerLogs(_ context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	if f.logFunc != nil {
		return f.logFunc(container, options)
	}
	return nil, nil
}

func (f *fakeClient) ClientVersion() string {
	return f.Version
}

func (f *fakeClient) ContainerWait(_ context.Context, container string, _ container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	if f.waitFunc != nil {
		return f.waitFunc(container)
	}
	return nil, nil
}

func (f *fakeClient) ContainerStart(_ context.Context, container string, options types.ContainerStartOptions) error {
	if f.containerStartFunc != nil {
		return f.containerStartFunc(container, options)
	}
	return nil
}

func (f *fakeClient) ContainerExport(_ context.Context, container string) (io.ReadCloser, error) {
	if f.containerExportFunc != nil {
		return f.containerExportFunc(container)
	}
	return nil, nil
}

func (f *fakeClient) ContainerExecResize(_ context.Context, id string, options types.ResizeOptions) error {
	if f.containerExecResizeFunc != nil {
		return f.containerExecResizeFunc(id, options)
	}
	return nil
}

func (f *fakeClient) ContainerKill(ctx context.Context, container, signal string) error {
	if f.containerKillFunc != nil {
		return f.containerKillFunc(ctx, container, signal)
	}
	return nil
}
