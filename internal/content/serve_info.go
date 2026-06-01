package content

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/monmsroutes"
)

// ServeURLs are operator-facing links shown at serve startup.
type ServeURLs struct {
	SiteURL      string
	AdminURL     string
	DashboardURL string
	PublishURL   string
	ConfigPath   string
	LogsDir      string
}

// ResolveServeURLs builds operator-facing links for the startup banner.
// When siteUrl is set in config, it is used as the base (no port).
// Otherwise URLs use the first allowedHost (or localhost) with the listen port.
func ResolveServeURLs(siteAbs string, serveArgs []string) (ServeURLs, error) {
	cfg, err := LoadMonmsConfig(siteAbs)
	if err != nil {
		return ServeURLs{}, err
	}

	base := resolveDisplayBase(cfg, serveArgs)

	configPath := filepath.Join(siteAbs, ".monms", "config.json")
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return ServeURLs{}, fmt.Errorf("resolve config path: %w", err)
	}

	return ServeURLs{
		SiteURL:      base + "/",
		AdminURL:     base + "/_/",
		DashboardURL: base + monmsroutes.DashboardHomePath,
		PublishURL:   base + monmsroutes.PublishPath,
		ConfigPath:   configPath,
		LogsDir:      filepath.Join(siteAbs, ".monms", "logs"),
	}, nil
}

func resolveDisplayBase(cfg MonmsConfig, serveArgs []string) string {
	if base := strings.TrimSpace(cfg.SiteURL); base != "" {
		return strings.TrimRight(base, "/")
	}
	host := displayHost(cfg.AllowedHosts)
	port := serveHTTPPort(serveArgs, cfg)
	return fmt.Sprintf("http://%s:%s", host, port)
}

// PrintServeBanner writes startup URLs for operators when stdout is a TTY.
func PrintServeBanner(urls ServeURLs, out io.Writer) {
	if out == nil {
		out = os.Stdout
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "MonMS is running\n")
	fmt.Fprintf(out, "  Site:      %s\n", urls.SiteURL)
	fmt.Fprintf(out, "  Dashboard: %s\n", urls.DashboardURL)
	fmt.Fprintf(out, "  Admin:     %s\n", urls.AdminURL)
	fmt.Fprintf(out, "  Publish:   %s\n", urls.PublishURL)
	fmt.Fprintf(out, "  Options:  edit %s\n", urls.ConfigPath)
	fmt.Fprintf(out, "  Logs:     %s\n", urls.LogsDir)
	fmt.Fprintln(out)
}

func displayHost(allowedHosts []string) string {
	for _, h := range normalizeAllowedHosts(allowedHosts) {
		return h
	}
	return "localhost"
}

func serveHTTPPort(serveArgs []string, cfg MonmsConfig) string {
	if host, port, ok := parseHTTPFlag(serveArgs); ok {
		_ = host
		if port != "" {
			return port
		}
	}
	if cfg.Bind != nil {
		if p := strings.TrimSpace(cfg.Bind.Port); p != "" {
			return p
		}
	}
	return "8090"
}

func parseHTTPFlag(args []string) (host, port string, ok bool) {
	for i, a := range args {
		if strings.HasPrefix(a, "--http=") {
			return splitHostPort(strings.TrimPrefix(a, "--http="))
		}
		if a == "--http" && i+1 < len(args) {
			return splitHostPort(args[i+1])
		}
	}
	return "", "", false
}

func splitHostPort(addr string) (host, port string, ok bool) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", "", false
	}
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(addr, ":") {
			return "", "", false
		}
		return addr, "8090", true
	}
	if p == "" {
		p = "8090"
	}
	return h, p, true
}
