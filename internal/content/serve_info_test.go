package content

import (
	"path/filepath"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestResolveServeURLsUsesSiteURL(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "siteUrl": "https://staging.example.com",
  "productionUrl": "https://production.example.com",
  "bind": { "host": "0.0.0.0", "port": "8090" },
  "allowedHosts": ["monms.example.com"]
}`)

	urls, err := ResolveServeURLs(ws, []string{"serve", "--http=0.0.0.0:8090"})
	if err != nil {
		t.Fatal(err)
	}
	if urls.SiteURL != "https://staging.example.com/" {
		t.Fatalf("SiteURL = %q", urls.SiteURL)
	}
	if urls.AdminURL != "https://staging.example.com/_/" {
		t.Fatalf("AdminURL = %q", urls.AdminURL)
	}
	if urls.DashboardURL != "https://staging.example.com/_monms/" {
		t.Fatalf("DashboardURL = %q", urls.DashboardURL)
	}
	if urls.PublishURL != "https://staging.example.com/_monms/publish" {
		t.Fatalf("PublishURL = %q", urls.PublishURL)
	}
}

func TestResolveServeURLsSiteURLStripsTrailingSlash(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "siteUrl": "https://staging.example.com/"
}`)

	urls, err := ResolveServeURLs(ws, nil)
	if err != nil {
		t.Fatal(err)
	}
	if urls.SiteURL != "https://staging.example.com/" {
		t.Fatalf("SiteURL = %q", urls.SiteURL)
	}
}

func TestResolveServeURLsIgnoresProductionURL(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "productionUrl": "https://production.example.com",
  "bind": { "host": "0.0.0.0", "port": "8090" },
  "allowedHosts": ["staging.example.com"]
}`)

	urls, err := ResolveServeURLs(ws, []string{"serve", "--http=0.0.0.0:8090"})
	if err != nil {
		t.Fatal(err)
	}
	if urls.SiteURL != "http://staging.example.com:8090/" {
		t.Fatalf("SiteURL = %q, want allowedHosts fallback not productionUrl", urls.SiteURL)
	}
}

func TestResolveServeURLsUsesAllowedHost(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "bind": { "host": "0.0.0.0", "port": "8090" },
  "allowedHosts": ["monms.example.com", "other.example.com"]
}`)

	urls, err := ResolveServeURLs(ws, []string{"serve", "--http=0.0.0.0:8090"})
	if err != nil {
		t.Fatal(err)
	}
	if urls.SiteURL != "http://monms.example.com:8090/" {
		t.Fatalf("SiteURL = %q", urls.SiteURL)
	}
	if urls.AdminURL != "http://monms.example.com:8090/_/" {
		t.Fatalf("AdminURL = %q", urls.AdminURL)
	}
	if !filepath.IsAbs(urls.ConfigPath) {
		t.Fatalf("ConfigPath not absolute: %q", urls.ConfigPath)
	}
}

func TestResolveServeURLsDefaultLocalhost(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, ".monms/config.json"), `{
  "bind": { "host": "0.0.0.0", "port": "9001" }
}`)

	urls, err := ResolveServeURLs(ws, []string{"serve", "--http=0.0.0.0:9001"})
	if err != nil {
		t.Fatal(err)
	}
	if urls.SiteURL != "http://localhost:9001/" {
		t.Fatalf("SiteURL = %q", urls.SiteURL)
	}
}

func TestDisplayHost(t *testing.T) {
	if got := displayHost(nil); got != "localhost" {
		t.Fatalf("got %q", got)
	}
	if got := displayHost([]string{"", "  ", "app.test"}); got != "app.test" {
		t.Fatalf("got %q", got)
	}
}
