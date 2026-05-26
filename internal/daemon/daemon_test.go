package daemon

import "testing"

func TestStripDaemonFlag(t *testing.T) {
	t.Parallel()

	clean, requested := StripDaemonFlag([]string{"serve", "-d", "--workspace", "ws"})
	if !requested {
		t.Fatal("expected daemon flag")
	}
	want := []string{"serve", "--workspace", "ws"}
	if len(clean) != len(want) {
		t.Fatalf("clean = %v, want %v", clean, want)
	}
	for i := range want {
		if clean[i] != want[i] {
			t.Fatalf("clean = %v, want %v", clean, want)
		}
	}
}

func TestShouldDetach(t *testing.T) {
	t.Parallel()

	if !ShouldDetach([]string{"serve", "-d"}) {
		t.Fatal("expected detach for serve -d")
	}
	if !ShouldDetach([]string{"-d"}) {
		t.Fatal("expected detach for implicit serve -d")
	}
	if ShouldDetach([]string{"serve"}) {
		t.Fatal("did not expect detach without -d")
	}
	if ShouldDetach([]string{"stop"}) {
		t.Fatal("did not expect detach for stop")
	}
	if ShouldDetach([]string{"init", "-d"}) {
		t.Fatal("did not expect detach for init -d")
	}
}

func TestServeRequested(t *testing.T) {
	t.Parallel()

	if !ServeRequested([]string{"serve"}) {
		t.Fatal("expected serve")
	}
	if ServeRequested([]string{"stop"}) {
		t.Fatal("did not expect serve")
	}
}
