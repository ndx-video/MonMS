package content

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestInstallerBaseURLUsesSiteURL(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "siteUrl": "https://monms.gandalf.lan",
  "bind": { "host": "0.0.0.0", "port": "8090" }
}`)

	got := installerBaseURL(ws, []string{"serve", "--http=0.0.0.0:8090"})
	if got != "https://monms.gandalf.lan" {
		t.Fatalf("got %q", got)
	}
}

func TestInstallerBaseURLFallsBackToAllowedHost(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "allowedHosts": ["staging.example.com"],
  "bind": { "host": "0.0.0.0", "port": "8090" }
}`)

	got := installerBaseURL(ws, []string{"serve", "--http=0.0.0.0:8090"})
	if got != "http://staging.example.com:8090" {
		t.Fatalf("got %q", got)
	}
}
