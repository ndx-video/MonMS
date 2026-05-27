package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultSite = "./site"

// ResolveSite parses -s/--site from args, then MONMS_SITE env,
// then defaults to "./site". Flag wins over env (D-26).
func ResolveSite(args []string, env []string) (configured string, absolute string, err error) {
	configured = defaultSite
	flagSet := false
	fromOverride := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if val, eq, ok := siteFlagValue(arg); ok && eq {
			configured = val
			flagSet = true
			fromOverride = true
			continue
		}
		if _, eq, ok := siteFlagValue(arg); ok && !eq {
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for --site flag")
			}
			configured = args[i+1]
			i++
			flagSet = true
			fromOverride = true
		}
	}

	if !flagSet {
		if v := envValue(env, "MONMS_SITE"); v != "" {
			configured = v
			fromOverride = true
		}
	}

	if fromOverride {
		configured = filepath.Clean(configured)
		if configured == "" || configured == "." {
			return "", "", fmt.Errorf("site path must not be empty")
		}
	}

	absolute, err = filepath.Abs(configured)
	if err != nil {
		return "", "", fmt.Errorf("resolve site absolute path: %w", err)
	}

	return configured, absolute, nil
}

// StripSiteFlags removes -s/--site and its value from args so downstream
// CLIs (e.g. PocketBase cobra) do not reject unknown flags.
func StripSiteFlags(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if val, eq, ok := siteFlagValue(arg); ok {
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

// siteFlagValue reports whether arg is -s or --site and an optional inline value.
func siteFlagValue(arg string) (value string, hasEqualsForm bool, ok bool) {
	switch {
	case strings.HasPrefix(arg, "--site="):
		return strings.TrimPrefix(arg, "--site="), true, true
	case strings.HasPrefix(arg, "-s="):
		return strings.TrimPrefix(arg, "-s="), true, true
	case arg == "--site", arg == "-s":
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
