//go:build !linux
// +build !linux

package archive // import "github.com/DevanshMathur19/docker-v23/pkg/archive"

func getWhiteoutConverter(format WhiteoutFormat, inUserNS bool) (tarWhiteoutConverter, error) {
	return nil, nil
}
