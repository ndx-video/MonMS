package content

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestApplyServeConfigFromWorkspace(t *testing.T) {
	t.Run("no config leaves args unchanged", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		got, err := ApplyServeConfigFromWorkspace(ws, []string{"serve"})
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		if len(got) != 1 || got[0] != "serve" {
			t.Fatalf("got %v, want [serve]", got)
		}
	})

	t.Run("bind injects http", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "bind": { "host": "0.0.0.0", "port": "8090" }
}`)
		got, err := ApplyServeConfigFromWorkspace(ws, []string{"serve"})
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		want := []string{"serve", "--http=0.0.0.0:8090"}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})

	t.Run("CLI http wins over bind config", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{"bind":{"host":"0.0.0.0","port":"8090"}}`)
		args := []string{"serve", "--http", "127.0.0.1:9090"}
		got, err := ApplyServeConfigFromWorkspace(ws, args)
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		if len(got) != len(args) {
			t.Fatalf("got %v, want unchanged %v", got, args)
		}
		for i := range args {
			if got[i] != args[i] {
				t.Fatalf("got %v, want %v", got, args)
			}
		}
	})

	t.Run("allowedHosts injects origins", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "allowedHosts": ["monms.gandalf.lan", "staging.example.com"]
}`)
		got, err := ApplyServeConfigFromWorkspace(ws, []string{"serve"})
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		want := []string{"serve", "--origins=monms.gandalf.lan,staging.example.com"}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})

	t.Run("bind and allowedHosts inject both flags", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "bind": { "host": "0.0.0.0", "port": "8090" },
  "allowedHosts": ["monms.gandalf.lan"]
}`)
		got, err := ApplyServeConfigFromWorkspace(ws, []string{"serve"})
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		want := []string{"serve", "--http=0.0.0.0:8090", "--origins=monms.gandalf.lan"}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})

	t.Run("CLI origins wins over config", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{"allowedHosts":["monms.gandalf.lan"]}`)
		args := []string{"serve", "--origins", "cli.example.com"}
		got, err := ApplyServeConfigFromWorkspace(ws, args)
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		if len(got) != len(args) {
			t.Fatalf("got %v, want unchanged %v", got, args)
		}
		for i := range args {
			if got[i] != args[i] {
				t.Fatalf("got %v, want %v", got, args)
			}
		}
	})

	t.Run("empty allowedHosts skipped", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{"allowedHosts":["", "  "]}`)
		got, err := ApplyServeConfigFromWorkspace(ws, []string{"serve"})
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %v, want [serve]", got)
		}
	})

	t.Run("invalid bind port skipped", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{"bind":{"host":"0.0.0.0","port":"nope"}}`)
		got, err := ApplyServeConfigFromWorkspace(ws, []string{"serve"})
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %v, want [serve]", got)
		}
	})

	t.Run("help flag skips injection", func(t *testing.T) {
		ws := testutil.NewWorkspace(t)
		testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "bind": { "host": "0.0.0.0", "port": "8090" },
  "allowedHosts": ["monms.gandalf.lan"]
}`)
		got, err := ApplyServeConfigFromWorkspace(ws, []string{"serve", "--help"})
		if err != nil {
			t.Fatalf("ApplyServeConfigFromWorkspace: %v", err)
		}
		if len(got) != 2 || got[0] != "serve" || got[1] != "--help" {
			t.Fatalf("got %v, want [serve --help]", got)
		}
	})
}

func TestBindAddress(t *testing.T) {
	tests := []struct {
		name string
		bind *BindConfig
		want string
		ok   bool
	}{
		{"nil", nil, "", false},
		{"empty", &BindConfig{}, "", false},
		{"host and port", &BindConfig{Host: "0.0.0.0", Port: "8090"}, "0.0.0.0:8090", true},
		{"port only defaults host", &BindConfig{Port: "9090"}, "127.0.0.1:9090", true},
		{"host only defaults port", &BindConfig{Host: "0.0.0.0"}, "0.0.0.0:8090", true},
		{"invalid port", &BindConfig{Host: "0.0.0.0", Port: "nope"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := bindAddress(tt.bind)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("bindAddress() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}
