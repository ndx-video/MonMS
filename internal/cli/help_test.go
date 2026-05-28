package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseHelpRequest(t *testing.T) {
	tests := []struct {
		args     []string
		cmd      string
		wantHelp bool
	}{
		{[]string{"--help"}, "", true},
		{[]string{"-h"}, "", true},
		{[]string{"init", "--help"}, "init", true},
		{[]string{"help"}, "", true},
		{[]string{"help", "init"}, "init", true},
		{[]string{"serve", "--help"}, "serve", true},
		{[]string{"init"}, "", false},
		{[]string{"content", "export"}, "", false},
	}
	for _, tc := range tests {
		cmd, ok := ParseHelpRequest(tc.args)
		if ok != tc.wantHelp || cmd != tc.cmd {
			t.Errorf("ParseHelpRequest(%v) = (%q, %v), want (%q, %v)", tc.args, cmd, ok, tc.cmd, tc.wantHelp)
		}
	}
}

func TestRootHelp(t *testing.T) {
	var buf bytes.Buffer
	printRootHelp(&buf)
	out := buf.String()
	for _, needle := range []string{"init", "validate", "content", "stop", "restart", "serve", "monms init"} {
		if !strings.Contains(out, needle) {
			t.Errorf("root help missing %q", needle)
		}
	}
}

func TestInitHelpRequest(t *testing.T) {
	cmd, wantHelp := ParseHelpRequest([]string{"init", "--help"})
	if !wantHelp || cmd != "init" || !IsMonmsCommand(cmd) {
		t.Fatalf("expected init help request, got cmd=%q wantHelp=%v", cmd, wantHelp)
	}
}

func TestContentSubcommandHelp(t *testing.T) {
	text, ok := ContentSubcommandHelp("export")
	if !ok || !strings.Contains(text, "content export") {
		t.Fatalf("export help: ok=%v text=%q", ok, text)
	}
}
