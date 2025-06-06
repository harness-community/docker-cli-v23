package image

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harness-community/docker-cli-v23/e2e/internal/fixtures"
	"github.com/harness-community/docker-cli-v23/internal/test/environment"
	"github.com/harness-community/docker-cli-v23/internal/test/output"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
	"gotest.tools/v3/skip"
)

func TestBuildFromContextDirectoryWithTag(t *testing.T) {
	t.Setenv("DOCKER_BUILDKIT", "0")

	dir := fs.NewDir(t, "test-build-context-dir",
		fs.WithFile("run", "echo running", fs.WithMode(0o755)),
		fs.WithDir("data", fs.WithFile("one", "1111")),
		fs.WithFile("Dockerfile", fmt.Sprintf(`
	FROM %s
	COPY run /usr/bin/run
	RUN run
	COPY data /data
		`, fixtures.AlpineImage)))
	defer dir.Remove()

	result := icmd.RunCmd(
		icmd.Command("docker", "build", "-t", "myimage", "."),
		withWorkingDir(dir))
	defer icmd.RunCommand("docker", "image", "rm", "myimage")

	const buildkitDisabledWarning = `DEPRECATED: The legacy builder is deprecated and will be removed in a future release.
            BuildKit is currently disabled; enable it by removing the DOCKER_BUILDKIT=0
            environment-variable.
`

	result.Assert(t, icmd.Expected{Err: buildkitDisabledWarning})
	output.Assert(t, result.Stdout(), map[int]func(string) error{
		0:  output.Prefix("Sending build context to Docker daemon"),
		1:  output.Suffix("Step 1/4 : FROM registry:5000/alpine:3.6"),
		3:  output.Suffix("Step 2/4 : COPY run /usr/bin/run"),
		5:  output.Suffix("Step 3/4 : RUN run"),
		7:  output.Suffix("running"),
		8:  output.Contains("Removing intermediate container"),
		10: output.Suffix("Step 4/4 : COPY data /data"),
		12: output.Contains("Successfully built "),
		13: output.Suffix("Successfully tagged myimage:latest"),
	})
}

func TestTrustedBuild(t *testing.T) {
	skip.If(t, environment.RemoteDaemon())
	t.Setenv("DOCKER_BUILDKIT", "0")

	dir := fixtures.SetupConfigFile(t)
	defer dir.Remove()
	image1 := fixtures.CreateMaskedTrustedRemoteImage(t, registryPrefix, "trust-build1", "latest")
	image2 := fixtures.CreateMaskedTrustedRemoteImage(t, registryPrefix, "trust-build2", "latest")

	buildDir := fs.NewDir(t, "test-trusted-build-context-dir",
		fs.WithFile("Dockerfile", fmt.Sprintf(`
	FROM %s as build-base
	RUN echo ok > /foo
	FROM %s
	COPY --from=build-base foo bar
		`, image1, image2)))
	defer buildDir.Remove()

	result := icmd.RunCmd(
		icmd.Command("docker", "build", "-t", "myimage", "."),
		withWorkingDir(buildDir),
		fixtures.WithConfig(dir.Path()),
		fixtures.WithTrust,
		fixtures.WithNotary,
	)

	result.Assert(t, icmd.Expected{
		Out: fmt.Sprintf("FROM %s@sha", image1[:len(image1)-7]),
		Err: fmt.Sprintf("Tagging %s@sha", image1[:len(image1)-7]),
	})
	result.Assert(t, icmd.Expected{
		Out: fmt.Sprintf("FROM %s@sha", image2[:len(image2)-7]),
	})
}

func TestTrustedBuildUntrustedImage(t *testing.T) {
	skip.If(t, environment.RemoteDaemon())
	t.Setenv("DOCKER_BUILDKIT", "0")

	dir := fixtures.SetupConfigFile(t)
	defer dir.Remove()
	buildDir := fs.NewDir(t, "test-trusted-build-context-dir",
		fs.WithFile("Dockerfile", fmt.Sprintf(`
	FROM %s
	RUN []
		`, fixtures.AlpineImage)))
	defer buildDir.Remove()

	result := icmd.RunCmd(
		icmd.Command("docker", "build", "-t", "myimage", "."),
		withWorkingDir(buildDir),
		fixtures.WithConfig(dir.Path()),
		fixtures.WithTrust,
		fixtures.WithNotary,
	)

	result.Assert(t, icmd.Expected{
		ExitCode: 1,
		Err:      "does not have trust data for",
	})
}

func TestBuildIidFileSquash(t *testing.T) {
	environment.SkipIfNotExperimentalDaemon(t)
	t.Setenv("DOCKER_BUILDKIT", "0")

	dir := fs.NewDir(t, "test-iidfile-squash")
	defer dir.Remove()
	iidfile := filepath.Join(dir.Path(), "idsquash")
	buildDir := fs.NewDir(t, "test-iidfile-squash-build",
		fs.WithFile("Dockerfile", fmt.Sprintf(`
	FROM %s
	ENV FOO=FOO
	ENV BAR=BAR
	RUN touch /fiip
	RUN touch /foop`, fixtures.AlpineImage)),
	)
	defer buildDir.Remove()

	imageTag := "testbuildiidfilesquash"
	result := icmd.RunCmd(
		icmd.Command("docker", "build", "--iidfile", iidfile, "--squash", "-t", imageTag, "."),
		withWorkingDir(buildDir),
	)
	result.Assert(t, icmd.Success)
	id, err := os.ReadFile(iidfile)
	assert.NilError(t, err)
	result = icmd.RunCommand("docker", "image", "inspect", "-f", "{{.Id}}", imageTag)
	result.Assert(t, icmd.Success)
	assert.Check(t, is.Equal(string(id), strings.TrimSpace(result.Combined())))
}

func withWorkingDir(dir *fs.Dir) func(*icmd.Cmd) {
	return func(cmd *icmd.Cmd) {
		cmd.Dir = dir.Path()
	}
}
