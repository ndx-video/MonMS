package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveWorkspaceFlag(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		env        []string
		wantConfig string
		wantAbs    string
		wantErr    bool
	}{
		{
			name:       "flag space separated",
			args:       []string{"--workspace", "/tmp/ws"},
			wantConfig: "/tmp/ws",
			wantAbs:    "/tmp/ws",
		},
		{
			name:       "flag equals form",
			args:       []string{"--workspace=/tmp/ws-eq"},
			wantConfig: "/tmp/ws-eq",
			wantAbs:    "/tmp/ws-eq",
		},
		{
			name:       "env when flag absent",
			args:       nil,
			env:        []string{"MONMS_WORKSPACE=/env/ws"},
			wantConfig: "/env/ws",
			wantAbs:    "/env/ws",
		},
		{
			name:       "flag wins over env",
			args:       []string{"--workspace", "/flag/ws"},
			env:        []string{"MONMS_WORKSPACE=/env/ws"},
			wantConfig: "/flag/ws",
			wantAbs:    "/flag/ws",
		},
		{
			name:       "default workspace",
			args:       nil,
			env:        nil,
			wantConfig: "./workspace",
		},
		{
			name:       "short flag space separated",
			args:       []string{"-w", "/tmp/ws-short"},
			wantConfig: "/tmp/ws-short",
			wantAbs:    "/tmp/ws-short",
		},
		{
			name:       "short flag equals form",
			args:       []string{"-w=/tmp/ws-short-eq"},
			wantConfig: "/tmp/ws-short-eq",
			wantAbs:    "/tmp/ws-short-eq",
		},
		{
			name:    "missing short flag value",
			args:    []string{"-w"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, gotAbs, err := ResolveWorkspace(tt.args, tt.env)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveWorkspace() error = %v", err)
			}
			if gotConfig != tt.wantConfig {
				t.Errorf("configured = %q, want %q", gotConfig, tt.wantConfig)
			}
			if tt.wantAbs != "" {
				if gotAbs != tt.wantAbs {
					t.Errorf("absolute = %q, want %q", gotAbs, tt.wantAbs)
				}
				return
			}
			if !strings.HasSuffix(gotAbs, filepath.Join("workspace")) && gotConfig == "./workspace" {
				if !strings.Contains(gotAbs, "workspace") {
					t.Errorf("absolute = %q, expected path ending in workspace", gotAbs)
				}
			}
		})
	}
}

func TestStripWorkspaceFlags(t *testing.T) {
	t.Parallel()

	got := StripWorkspaceFlags([]string{"serve", "-w", "ws", "--dev"})
	want := []string{"serve", "--dev"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}

	got2 := StripWorkspaceFlags([]string{"content", "export", "--workspace=./ws"})
	want2 := []string{"content", "export"}
	if len(got2) != len(want2) || got2[0] != want2[0] || got2[1] != want2[1] {
		t.Fatalf("got %v, want %v", got2, want2)
	}
}
