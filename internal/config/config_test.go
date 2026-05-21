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
			name:    "missing flag value",
			args:    []string{"--workspace"},
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
