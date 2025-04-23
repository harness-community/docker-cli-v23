package build

import (
	"path/filepath"

	"github.com/DevanshMathur19/docker-v23/pkg/longpath"
)

func getContextRoot(srcPath string) (string, error) {
	cr, err := filepath.Abs(srcPath)
	if err != nil {
		return "", err
	}
	return longpath.AddPrefix(cr), nil
}
