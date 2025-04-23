//go:build !windows
// +build !windows

package archive // import "github.com/DevanshMathur19/docker-v23/pkg/archive"

import (
	"path/filepath"
)

func normalizePath(path string) string {
	return filepath.ToSlash(path)
}
