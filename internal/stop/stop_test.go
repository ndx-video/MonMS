package stop

import "testing"

func TestSameExecutable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b string
		want bool
	}{
		{"/tmp/monms", "/tmp/monms", true},
		{"/tmp/monms (deleted)", "/tmp/monms", true},
		{"/tmp/monms", "/other/monms", false},
	}

	for _, tc := range tests {
		if got := sameExecutable(tc.a, tc.b); got != tc.want {
			t.Errorf("sameExecutable(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNormalizeProcExe(t *testing.T) {
	t.Parallel()

	got := normalizeProcExe("/usr/bin/monms (deleted)")
	if got != "/usr/bin/monms" {
		t.Fatalf("normalizeProcExe = %q", got)
	}
}

func TestRunCLIRejectsArgs(t *testing.T) {
	t.Parallel()

	if err := RunCLI([]string{"--force"}); err == nil {
		t.Fatal("expected error for unexpected args")
	}
}
