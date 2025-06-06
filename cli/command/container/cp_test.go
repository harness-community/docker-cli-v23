package container

import (
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/pkg/archive"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/skip"
)

func TestRunCopyWithInvalidArguments(t *testing.T) {
	testcases := []struct {
		doc         string
		options     copyOptions
		expectedErr string
	}{
		{
			doc: "copy between container",
			options: copyOptions{
				source:      "first:/path",
				destination: "second:/path",
			},
			expectedErr: "copying between containers is not supported",
		},
		{
			doc: "copy without a container",
			options: copyOptions{
				source:      "./source",
				destination: "./dest",
			},
			expectedErr: "must specify at least one container source",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.doc, func(t *testing.T) {
			err := runCopy(test.NewFakeCli(nil), testcase.options)
			assert.Error(t, err, testcase.expectedErr)
		})
	}
}

func TestRunCopyFromContainerToStdout(t *testing.T) {
	tarContent := "the tar content"

	fakeClient := &fakeClient{
		containerCopyFromFunc: func(container, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
			assert.Check(t, is.Equal("container", container))
			return io.NopCloser(strings.NewReader(tarContent)), types.ContainerPathStat{}, nil
		},
	}
	options := copyOptions{source: "container:/path", destination: "-"}
	cli := test.NewFakeCli(fakeClient)
	err := runCopy(cli, options)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(tarContent, cli.OutBuffer().String()))
	assert.Check(t, is.Equal("", cli.ErrBuffer().String()))
}

func TestRunCopyFromContainerToFilesystem(t *testing.T) {
	destDir := fs.NewDir(t, "cp-test",
		fs.WithFile("file1", "content\n"))
	defer destDir.Remove()

	fakeClient := &fakeClient{
		containerCopyFromFunc: func(container, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
			assert.Check(t, is.Equal("container", container))
			readCloser, err := archive.TarWithOptions(destDir.Path(), &archive.TarOptions{})
			return readCloser, types.ContainerPathStat{}, err
		},
	}
	options := copyOptions{source: "container:/path", destination: destDir.Path(), quiet: true}
	cli := test.NewFakeCli(fakeClient)
	err := runCopy(cli, options)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("", cli.OutBuffer().String()))
	assert.Check(t, is.Equal("", cli.ErrBuffer().String()))

	content, err := os.ReadFile(destDir.Join("file1"))
	assert.NilError(t, err)
	assert.Check(t, is.Equal("content\n", string(content)))
}

func TestRunCopyFromContainerToFilesystemMissingDestinationDirectory(t *testing.T) {
	destDir := fs.NewDir(t, "cp-test",
		fs.WithFile("file1", "content\n"))
	defer destDir.Remove()

	fakeClient := &fakeClient{
		containerCopyFromFunc: func(container, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
			assert.Check(t, is.Equal("container", container))
			readCloser, err := archive.TarWithOptions(destDir.Path(), &archive.TarOptions{})
			return readCloser, types.ContainerPathStat{}, err
		},
	}

	options := copyOptions{
		source:      "container:/path",
		destination: destDir.Join("missing", "foo"),
	}
	cli := test.NewFakeCli(fakeClient)
	err := runCopy(cli, options)
	assert.ErrorContains(t, err, destDir.Join("missing"))
}

func TestRunCopyToContainerFromFileWithTrailingSlash(t *testing.T) {
	srcFile := fs.NewFile(t, t.Name())
	defer srcFile.Remove()

	options := copyOptions{
		source:      srcFile.Path() + string(os.PathSeparator),
		destination: "container:/path",
	}
	cli := test.NewFakeCli(&fakeClient{})
	err := runCopy(cli, options)

	expectedError := "not a directory"
	if runtime.GOOS == "windows" {
		expectedError = "The filename, directory name, or volume label syntax is incorrect"
	}
	assert.ErrorContains(t, err, expectedError)
}

func TestRunCopyToContainerSourceDoesNotExist(t *testing.T) {
	options := copyOptions{
		source:      "/does/not/exist",
		destination: "container:/path",
	}
	cli := test.NewFakeCli(&fakeClient{})
	err := runCopy(cli, options)
	expected := "no such file or directory"
	if runtime.GOOS == "windows" {
		expected = "cannot find the file specified"
	}
	assert.ErrorContains(t, err, expected)
}

func TestSplitCpArg(t *testing.T) {
	testcases := []struct {
		doc               string
		path              string
		os                string
		expectedContainer string
		expectedPath      string
	}{
		{
			doc:          "absolute path with colon",
			os:           "linux",
			path:         "/abs/path:withcolon",
			expectedPath: "/abs/path:withcolon",
		},
		{
			doc:          "relative path with colon",
			path:         "./relative:path",
			expectedPath: "./relative:path",
		},
		{
			doc:          "absolute path with drive",
			os:           "windows",
			path:         `d:\abs\path`,
			expectedPath: `d:\abs\path`,
		},
		{
			doc:          "no separator",
			path:         "relative/path",
			expectedPath: "relative/path",
		},
		{
			doc:               "with separator",
			path:              "container:/opt/foo",
			expectedPath:      "/opt/foo",
			expectedContainer: "container",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.doc, func(t *testing.T) {
			skip.If(t, testcase.os != "" && testcase.os != runtime.GOOS)

			container, path := splitCpArg(testcase.path)
			assert.Check(t, is.Equal(testcase.expectedContainer, container))
			assert.Check(t, is.Equal(testcase.expectedPath, path))
		})
	}
}

func TestRunCopyFromContainerToFilesystemIrregularDestination(t *testing.T) {
	options := copyOptions{source: "container:/dev/null", destination: "/dev/random"}
	cli := test.NewFakeCli(nil)
	err := runCopy(cli, options)
	assert.Assert(t, err != nil)
	expected := `"/dev/random" must be a directory or a regular file`
	assert.ErrorContains(t, err, expected)
}
