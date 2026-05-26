package content

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// MonmsConfig is staging workspace config (workspace/.monms/config.json).
type MonmsConfig struct {
	ProductionURL   string      `json:"productionUrl"`
	PublisherEmails []string    `json:"publisherEmails"`
	AllowedHosts    []string    `json:"allowedHosts"`
	Bind            *BindConfig `json:"bind,omitempty"`
}

// BindConfig is the workspace listen address for monms serve (maps to PocketBase --http).
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

// LoadMonmsConfig reads workspace/.monms/config.json; missing file returns zero config.
func LoadMonmsConfig(wsAbs string) (MonmsConfig, error) {
	path := filepath.Join(wsAbs, ".monms", "config.json")
	if err := ensureUnderWorkspace(wsAbs, path); err != nil {
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
	return cfg, nil
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
