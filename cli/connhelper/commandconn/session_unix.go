//go:build !windows
// +build !windows

package commandconn

import (
	"os/exec"
)

func createSession(cmd *exec.Cmd) {
	// for supporting ssh connection helper with ProxyCommand
	// https://github.com/DevanshMathur19/docker-cli-v23/issues/1707
	cmd.SysProcAttr.Setsid = true
}
