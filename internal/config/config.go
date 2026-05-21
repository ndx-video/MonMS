package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultWorkspace = "./workspace"

// ResolveWorkspace parses --workspace from args, then MONMS_WORKSPACE env,
// then defaults to "./workspace". Flag wins over env (D-26).
func ResolveWorkspace(args []string, env []string) (configured string, absolute string, err error) {
	configured = defaultWorkspace
	flagSet := false
	fromOverride := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--workspace=") {
			configured = strings.TrimPrefix(arg, "--workspace=")
			flagSet = true
			fromOverride = true
			continue
		}
		if arg == "--workspace" {
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for --workspace flag")
			}
			configured = args[i+1]
			i++
			flagSet = true
			fromOverride = true
		}
	}

	if !flagSet {
		if v := envValue(env, "MONMS_WORKSPACE"); v != "" {
			configured = v
			fromOverride = true
		}
	}

	if fromOverride {
		configured = filepath.Clean(configured)
		if configured == "" || configured == "." {
			return "", "", fmt.Errorf("workspace path must not be empty")
		}
	}

	absolute, err = filepath.Abs(configured)
	if err != nil {
		return "", "", fmt.Errorf("resolve workspace absolute path: %w", err)
	}

	return configured, absolute, nil
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	if env == nil {
		return os.Getenv(key)
	}
	return ""
}
