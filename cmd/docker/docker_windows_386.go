//go:build windows && 386
// +build windows,386

//go:generate goversioninfo -o=../../cli/winresources/resource.syso -icon=winresources/docker.ico -manifest=winresources/docker.exe.manifest ../../cli/winresources/versioninfo.json

package main

import _ "github.com/DevanshMathur19/docker-cli-v23/cli/winresources"
