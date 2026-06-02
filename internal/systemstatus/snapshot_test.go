package systemstatus

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestSnapshotBasic(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{"siteUrl":"http://staging.test","mcp":{"enabled":true}}`)

	v, err := Snapshot(ws, []string{"serve", "--http=127.0.0.1:8090"}, "development", true)
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if !strings.HasSuffix(filepath.Clean(v.SiteAbs), filepath.Clean(ws)) {
		t.Fatalf("SiteAbs %q", v.SiteAbs)
	}
	if v.BuildMode != "development" {
		t.Fatalf("BuildMode %q", v.BuildMode)
	}
	if v.HTTPListen != "127.0.0.1:8090" {
		t.Fatalf("HTTPListen %q", v.HTTPListen)
	}
	if !v.MCP.Enabled || v.MCPListen == "" {
		t.Fatalf("MCP %+v listen %q", v.MCP, v.MCPListen)
	}
	if !v.PublishTokenSet {
		t.Fatal("expected publish token set")
	}
}
