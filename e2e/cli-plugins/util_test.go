package cliplugins

import (
	"fmt"
	"os"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/config"
	"github.com/harness-community/docker-cli-v23/cli/config/configfile"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

func prepare(t *testing.T) (func(args ...string) icmd.Cmd, *configfile.ConfigFile, func()) {
	cfg := fs.NewDir(t, "plugin-test",
		fs.WithFile("config.json", fmt.Sprintf(`{"cliPluginsExtraDirs": [%q]}`, os.Getenv("DOCKER_CLI_E2E_PLUGINS_EXTRA_DIRS"))),
	)
	run := func(args ...string) icmd.Cmd {
		return icmd.Command("docker", append([]string{"--config", cfg.Path()}, args...)...)
	}
	cleanup := func() {
		cfg.Remove()
	}
	cfgfile, err := config.Load(cfg.Path())
	assert.NilError(t, err)

	return run, cfgfile, cleanup
}
