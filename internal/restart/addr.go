package restart

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

// IsAddrInUse reports whether err is a TCP bind/listen "address already in use" failure.
func IsAddrInUse(err error) bool {
	if err == nil {
		return false
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var errno syscall.Errno
		if errors.As(opErr.Err, &errno) && errno == syscall.EADDRINUSE {
			return true
		}
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "address already in use") ||
		strings.Contains(msg, "only one usage of each socket address")
}
