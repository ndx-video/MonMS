package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultWorkspace = "./workspace"

// ResolveWorkspace parses -w/--workspace from args, then MONMS_WORKSPACE env,
// then defaults to "./workspace". Flag wins over env (D-26).
func ResolveWorkspace(args []string, env []string) (configured string, absolute string, err error) {
	configured = defaultWorkspace
	flagSet := false
	fromOverride := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if val, eq, ok := workspaceFlagValue(arg); ok && eq {
			configured = val
			flagSet = true
			fromOverride = true
			continue
		}
		if _, eq, ok := workspaceFlagValue(arg); ok && !eq {
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

// StripWorkspaceFlags removes -w/--workspace and its value from args so downstream
// CLIs (e.g. PocketBase cobra) do not reject unknown flags.
func StripWorkspaceFlags(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if val, eq, ok := workspaceFlagValue(arg); ok {
			if !eq && val == "" {
				if i+1 < len(args) {
					i++
				}
			}
			continue
		}
		out = append(out, arg)
	}
	return out
}

// workspaceFlagValue reports whether arg is -w or --workspace and an optional inline value.
func workspaceFlagValue(arg string) (value string, hasEqualsForm bool, ok bool) {
	switch {
	case strings.HasPrefix(arg, "--workspace="):
		return strings.TrimPrefix(arg, "--workspace="), true, true
	case strings.HasPrefix(arg, "-w="):
		return strings.TrimPrefix(arg, "-w="), true, true
	case arg == "--workspace", arg == "-w":
		return "", false, true
	default:
		return "", false, false
	}
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
