package content

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/monms/monms/internal/cli"
)

// ApplyServeConfigFromSite injects PocketBase serve flags from site config
// when the CLI did not already specify them. CLI flags win over config (D-26 parity).
func ApplyServeConfigFromSite(siteAbs string, args []string) ([]string, error) {
	if cli.HasHelpFlag(args) {
		return args, nil
	}

	cfg, err := LoadMonmsConfig(siteAbs)
	if err != nil {
		return nil, err
	}

	out := append([]string(nil), args...)

	if !hasHTTPFlag(args) {
		if addr, ok := bindAddress(cfg.Bind); ok {
			slog.Info("serve bind from site config", "http", addr)
			out = append(out, "--http="+addr)
		}
	}

	if !hasOriginsFlag(args) {
		hosts := normalizeAllowedHosts(cfg.AllowedHosts)
		if len(hosts) > 0 {
			slog.Info("serve origins from site config", "hosts", hosts)
			out = append(out, "--origins="+strings.Join(hosts, ","))
		}
	}

	return out, nil
}

func bindAddress(bind *BindConfig) (string, bool) {
	if bind == nil {
		return "", false
	}

	host := strings.TrimSpace(bind.Host)
	port := strings.TrimSpace(bind.Port)
	if host == "" && port == "" {
		return "", false
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "8090"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", false
	}

	return fmt.Sprintf("%s:%d", host, n), true
}

func hasHTTPFlag(args []string) bool {
	for _, a := range args {
		if a == "--http" || strings.HasPrefix(a, "--http=") {
			return true
		}
	}
	return false
}

func hasOriginsFlag(args []string) bool {
	for _, a := range args {
		if a == "--origins" || strings.HasPrefix(a, "--origins=") {
			return true
		}
	}
	return false
}

func normalizeAllowedHosts(hosts []string) []string {
	if len(hosts) == 0 {
		return nil
	}
	out := make([]string, 0, len(hosts))
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if h != "" {
			out = append(out, h)
		}
	}
	return out
}
