package content

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monms/monms/internal/site"
	"github.com/pocketbase/pocketbase/core"
)

// MCPConfig controls the optional MonMS MCP HTTP listener and API key policy.
type MCPConfig struct {
	Enabled               bool   `json:"enabled"`
	Host                  string `json:"host"`
	Port                  string `json:"port"`
	AllowNonSuperuserKeys bool   `json:"allowNonSuperuserKeys"`
}

// DefaultMCPConfig returns safe defaults for a new site.
func DefaultMCPConfig() MCPConfig {
	return MCPConfig{
		Enabled:               false,
		Host:                  "127.0.0.1",
		Port:                  "8091",
		AllowNonSuperuserKeys: false,
	}
}

// MonmsConfig is staging site config (site/.monms/config.json).
type MonmsConfig struct {
	ProductionURL     string                   `json:"productionUrl"`
	SiteURL           string                   `json:"siteUrl"`
	PublisherEmails   []string                 `json:"publisherEmails"`
	AllowedHosts      []string                 `json:"allowedHosts"`
	Bind              *BindConfig              `json:"bind,omitempty"`
	MCP               MCPConfig                `json:"mcp"`
	ShapeSync         *site.ShapeSyncConfig    `json:"shapeSync,omitempty"`
	Logging           []string                 `json:"logging,omitempty"`
	LoggingRotation   *LoggingRotationConfig   `json:"loggingRotation,omitempty"`
}

// LoggingRotationConfig controls lumberjack rotation for site log files.
type LoggingRotationConfig struct {
	MaxSizeMB  int  `json:"maxSizeMB"`
	MaxBackups int  `json:"maxBackups"`
	MaxAgeDays int  `json:"maxAgeDays"`
	Compress   bool `json:"compress"`
}

// BindConfig is the site listen address for monms serve (maps to PocketBase --http).
type BindConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// RequirePublishToken gates content import routes with MONMS_PUBLISH_TOKEN (PUB-05).
// Fails closed when expected is empty (production token unset).
func RequirePublishToken(expected string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if expected == "" {
			return e.UnauthorizedError("invalid publish token", nil)
		}

		auth := e.Request.Header.Get("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			return e.UnauthorizedError("invalid publish token", nil)
		}
		return e.Next()
	}
}

// LoadMonmsConfig reads site/.monms/config.json; missing file returns zero config.
func LoadMonmsConfig(siteAbs string) (MonmsConfig, error) {
	path := filepath.Join(siteAbs, ".monms", "config.json")
	if err := ensureUnderSite(siteAbs, path); err != nil {
		return MonmsConfig{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return MonmsConfig{}, nil
		}
		return MonmsConfig{}, fmt.Errorf("monms config: read: %w", err)
	}
	if len(data) == 0 {
		return MonmsConfig{}, nil
	}

	var cfg MonmsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return MonmsConfig{}, fmt.Errorf("monms config: parse: %w", err)
	}
	if cfg.MCP.Host == "" && cfg.MCP.Port == "" && !cfg.MCP.Enabled && !cfg.MCP.AllowNonSuperuserKeys {
		cfg.MCP = DefaultMCPConfig()
	} else {
		if cfg.MCP.Host == "" {
			cfg.MCP.Host = DefaultMCPConfig().Host
		}
		if cfg.MCP.Port == "" {
			cfg.MCP.Port = DefaultMCPConfig().Port
		}
	}
	return cfg, nil
}

// SaveMonmsMCPSettings updates the mcp block in site/.monms/config.json, preserving other keys.
func SaveMonmsMCPSettings(siteAbs string, mcp MCPConfig) error {
	path := filepath.Join(siteAbs, ".monms", "config.json")
	if err := ensureUnderSite(siteAbs, path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("monms config: create dir: %w", err)
	}

	doc := map[string]json.RawMessage{}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("monms config: read: %w", err)
	}
	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("monms config: parse: %w", err)
		}
	}

	mcpJSON, err := json.Marshal(mcp)
	if err != nil {
		return fmt.Errorf("monms config: encode mcp: %w", err)
	}
	doc["mcp"] = mcpJSON

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("monms config: encode: %w", err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("monms config: write: %w", err)
	}
	return nil
}

// IsPublisher reports whether email is in the publisher allowlist (PUB-07).
func IsPublisher(email string, allowed []string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	for _, a := range allowed {
		if email == strings.ToLower(strings.TrimSpace(a)) {
			return true
		}
	}
	return false
}

// RequirePublisher gates publish routes to allowlisted superuser emails (PUB-07).
func RequirePublisher(allowedEmails []string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.UnauthorizedError("authentication required", nil)
		}
		email := e.Auth.GetString("email")
		if !IsPublisher(email, allowedEmails) {
			return e.ForbiddenError("publisher role required", nil)
		}
		return e.Next()
	}
}