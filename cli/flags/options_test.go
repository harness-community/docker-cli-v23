package flags

import (
	"path/filepath"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/config"
	"github.com/spf13/pflag"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestClientOptionsInstallFlags(t *testing.T) {
	flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)
	opts := NewClientOptions()
	opts.InstallFlags(flags)

	err := flags.Parse([]string{
		"--tlscacert=\"/foo/cafile\"",
		"--tlscert=\"/foo/cert\"",
		"--tlskey=\"/foo/key\"",
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal("/foo/cafile", opts.TLSOptions.CAFile))
	assert.Check(t, is.Equal("/foo/cert", opts.TLSOptions.CertFile))
	assert.Check(t, is.Equal(opts.TLSOptions.KeyFile, "/foo/key"))
}

func defaultPath(filename string) string {
	return filepath.Join(config.Dir(), filename)
}

func TestClientOptionsInstallFlagsWithDefaults(t *testing.T) {
	flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)
	opts := NewClientOptions()
	opts.InstallFlags(flags)

	err := flags.Parse([]string{})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(defaultPath("ca.pem"), opts.TLSOptions.CAFile))
	assert.Check(t, is.Equal(defaultPath("cert.pem"), opts.TLSOptions.CertFile))
	assert.Check(t, is.Equal(defaultPath("key.pem"), opts.TLSOptions.KeyFile))
}
