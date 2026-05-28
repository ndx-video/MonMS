package restart

import (
	"errors"
	"net"
	"syscall"
	"testing"
)

func TestIsAddrInUse(t *testing.T) {
	t.Parallel()

	if IsAddrInUse(nil) {
		t.Fatal("nil error should be false")
	}
	if !IsAddrInUse(errors.New("listen tcp 127.0.0.1:8090: bind: address already in use")) {
		t.Fatal("expected string match")
	}

	opErr := &net.OpError{
		Op:  "listen",
		Err: syscall.EADDRINUSE,
	}
	if !IsAddrInUse(opErr) {
		t.Fatal("expected EADDRINUSE match")
	}
}

func TestRunCLIRejectsHelp(t *testing.T) {
	t.Parallel()

	if err := RunCLI([]string{"--help"}); err != nil {
		t.Fatalf("help: %v", err)
	}
}
