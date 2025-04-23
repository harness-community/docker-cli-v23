//go:build !linux
// +build !linux

package archive // import "github.com/harness-community/docker-v23/pkg/archive"

func getWhiteoutConverter(format WhiteoutFormat, inUserNS bool) (tarWhiteoutConverter, error) {
	return nil, nil
}
