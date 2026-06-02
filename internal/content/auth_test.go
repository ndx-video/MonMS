package content

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsPublisherCaseInsensitive(t *testing.T) {
	allowed := []string{"Publisher@Client.com", " editor@example.com "}

	if !IsPublisher("publisher@client.com", allowed) {
		t.Fatal("expected case-insensitive match for publisher email")
	}
	if !IsPublisher("  editor@example.com", allowed) {
		t.Fatal("expected trimmed case-insensitive match")
	}
	if IsPublisher("other@example.com", allowed) {
		t.Fatal("unexpected match for non-publisher email")
	}
}

func TestLoadMonmsConfigMCPMissingBlock(t *testing.T) {
	site := t.TempDir()
	writeConfig(t, site, `{"productionUrl":"https://example.com"}`)

	cfg, err := LoadMonmsConfig(site)
	if err != nil {
		t.Fatalf("LoadMonmsConfig: %v", err)
	}
	def := DefaultMCPConfig()
	if cfg.MCP != def {
		t.Fatalf("MCP = %+v, want full defaults %+v when mcp key absent", cfg.MCP, def)
	}
}

func TestLoadMonmsConfigMCPPartialBlock(t *testing.T) {
	site := t.TempDir()
	writeConfig(t, site, `{"mcp":{"enabled":true}}`)

	cfg, err := LoadMonmsConfig(site)
	if err != nil {
		t.Fatalf("LoadMonmsConfig: %v", err)
	}
	if !cfg.MCP.Enabled {
		t.Fatal("expected enabled=true from partial mcp block")
	}
	if cfg.MCP.Host != DefaultMCPConfig().Host {
		t.Fatalf("Host = %q, want default %q", cfg.MCP.Host, DefaultMCPConfig().Host)
	}
	if cfg.MCP.Port != DefaultMCPConfig().Port {
		t.Fatalf("Port = %q, want default %q", cfg.MCP.Port, DefaultMCPConfig().Port)
	}
	if cfg.MCP.AllowNonSuperuserKeys {
		t.Fatal("AllowNonSuperuserKeys should remain false when omitted")
	}
}

func TestLoadMonmsConfigMCPExplicitValues(t *testing.T) {
	site := t.TempDir()
	writeConfig(t, site, `{"mcp":{"enabled":true,"host":"0.0.0.0","port":"9000","allowNonSuperuserKeys":true}}`)

	cfg, err := LoadMonmsConfig(site)
	if err != nil {
		t.Fatalf("LoadMonmsConfig: %v", err)
	}
	if !cfg.MCP.Enabled || cfg.MCP.Host != "0.0.0.0" || cfg.MCP.Port != "9000" || !cfg.MCP.AllowNonSuperuserKeys {
		t.Fatalf("MCP = %+v, want explicit values preserved", cfg.MCP)
	}
}

func writeConfig(t *testing.T, siteAbs, body string) {
	t.Helper()
	dir := filepath.Join(siteAbs, ".monms")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}
