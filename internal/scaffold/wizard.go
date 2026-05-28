package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monms/monms/internal/cli/prompt"
)

// StartMode is the operator's choice after interactive setup.
type StartMode int

const (
	StartUnset StartMode = iota
	StartForeground
	StartBackground
	StartNone
)

const (
	defaultPort  = "8090"
	defaultBind  = "0.0.0.0"
	defaultHosts = "localhost"
)

// RunSetupWizard prompts for listen settings and writes site/.monms/config.json.
func RunSetupWizard(siteAbs string, p *prompt.Prompter) (StartMode, error) {
	portStr, err := p.ReadDefault("Specify the default listen port", defaultPort)
	if err != nil {
		return StartNone, err
	}
	port, err := parsePort(portStr)
	if err != nil {
		return StartNone, err
	}

	hostsInput, err := p.ReadDefault("Allowed hosts (space-separated list)", defaultHosts)
	if err != nil {
		return StartNone, err
	}
	allowedHosts := parseAllowedHosts(hostsInput)

	siteURL, err := p.ReadDefault(
		"Public site URL for this instance (optional, no trailing slash)",
		defaultSiteURL(allowedHosts, existingSiteURL(siteAbs)),
	)
	if err != nil {
		return StartNone, err
	}
	siteURL = strings.TrimRight(strings.TrimSpace(siteURL), "/")

	bindHost, err := p.ReadDefault("Bind host", defaultBind)
	if err != nil {
		return StartNone, err
	}
	bindHost = strings.TrimSpace(bindHost)
	if bindHost == "" {
		bindHost = defaultBind
	}

	if err := SaveMonmsServeSettings(siteAbs, siteURL, ServeBindConfig{
		Host: bindHost,
		Port: strconv.Itoa(port),
	}, allowedHosts); err != nil {
		return StartNone, err
	}

	idx, err := p.ReadChoice("How would you like to start the server?", []string{
		"foreground",
		"background",
		"exit without starting",
	}, 0)
	if err != nil {
		return StartNone, err
	}
	switch idx {
	case 1:
		return StartBackground, nil
	case 2:
		return StartNone, nil
	default:
		return StartForeground, nil
	}
}

func defaultSiteURL(allowedHosts []string, existing string) string {
	if len(allowedHosts) > 0 {
		if host := strings.TrimSpace(allowedHosts[0]); host != "" {
			return "https://" + host
		}
	}
	return strings.TrimRight(strings.TrimSpace(existing), "/")
}

func existingSiteURL(siteAbs string) string {
	path := filepath.Join(siteAbs, ".monms", "config.json")
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return ""
	}
	var doc struct {
		SiteURL string `json:"siteUrl"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return ""
	}
	return strings.TrimSpace(doc.SiteURL)
}

func parsePort(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		s = defaultPort
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return 0, fmt.Errorf("invalid port %q: must be 1-65535", s)
	}
	return n, nil
}

func parseAllowedHosts(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" || strings.EqualFold(input, defaultHosts) {
		return []string{"localhost", "127.0.0.1"}
	}
	var out []string
	for _, h := range strings.Fields(input) {
		h = strings.TrimSpace(h)
		if h != "" {
			out = append(out, h)
		}
	}
	return out
}
