package archive // import "github.com/DevanshMathur19/docker-v23/pkg/archive"

import (
	"syscall"
	"time"
)

func timeToTimespec(time time.Time) (ts syscall.Timespec) {
	if time.IsZero() {
		// Return UTIME_OMIT special value
		ts.Sec = 0
		ts.Nsec = (1 << 30) - 2
		return
	}
	return syscall.NsecToTimespec(time.UnixNano())
}
