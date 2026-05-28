package logging

import (
	"log/slog"
	"strings"
)

// LevelSchema is a custom slog level for MonMS schema bootstrap/sync events.
const LevelSchema slog.Level = 12

// LevelFlag selects per-level log files.
type LevelFlag uint8

const (
	FlagError LevelFlag = 1 << iota
	FlagWarn
	FlagInfo
	FlagDebug
	FlagSchema
)

const defaultFlags = FlagError

// AllLevelNames lists every configurable file log level (ERROR is always on).
var AllLevelNames = []string{"error", "warn", "info", "debug", "schema"}

var productionBuild bool

// SetProductionBuild selects build-mode defaults when config.json omits logging.
// Call from main before Configure (production ldflag → error, warn, schema; development → all levels).
func SetProductionBuild(production bool) {
	productionBuild = production
}

// DefaultLevelNames returns the logging array used when config.json has no logging key.
func DefaultLevelNames() []string {
	if productionBuild {
		return []string{"error", "warn", "schema"}
	}
	return append([]string(nil), AllLevelNames...)
}

// ParseLevels maps config strings to a level bitmask. ERROR is always enabled.
func ParseLevels(names []string) LevelFlag {
	flags := defaultFlags
	for _, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "error":
			flags |= FlagError
		case "warn", "warning":
			flags |= FlagWarn
		case "info":
			flags |= FlagInfo
		case "debug":
			flags |= FlagDebug
		case "schema":
			flags |= FlagSchema
		}
	}
	return flags
}

func (f LevelFlag) Has(flag LevelFlag) bool {
	return f&flag != 0
}

func slogLevelFlag(level slog.Level) LevelFlag {
	switch {
	case level == LevelSchema:
		return FlagSchema
	case level >= slog.LevelError:
		return FlagError
	case level >= slog.LevelWarn:
		return FlagWarn
	case level >= slog.LevelInfo:
		return FlagInfo
	default:
		return FlagDebug
	}
}

func levelName(level slog.Level) string {
	switch {
	case level == LevelSchema:
		return "SCHEMA"
	case level >= slog.LevelError:
		return "ERROR"
	case level >= slog.LevelWarn:
		return "WARN"
	case level >= slog.LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}
