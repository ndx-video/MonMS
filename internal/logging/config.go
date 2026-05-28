package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RotationConfig controls lumberjack rotation per log file.
type RotationConfig struct {
	MaxSizeMB  int  `json:"maxSizeMB"`
	MaxBackups int  `json:"maxBackups"`
	MaxAgeDays int  `json:"maxAgeDays"`
	Compress   bool `json:"compress"`
}

// Config is the resolved file logging layout for a site.
type Config struct {
	Dir       string
	Flags     LevelFlag
	Rotation  RotationConfig
	Files     *FileSet
}

type monmsConfigFile struct {
	Logging         []string          `json:"logging"`
	LoggingRotation *RotationConfig   `json:"loggingRotation"`
}

// LoadConfig reads logging settings from site/.monms/config.json.
func LoadConfig(siteAbs string) (Config, error) {
	path := filepath.Join(siteAbs, ".monms", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ConfigFromValues(siteAbs, nil, nil), nil
		}
		return Config{}, fmt.Errorf("logging config: read: %w", err)
	}
	if len(data) == 0 {
		return ConfigFromValues(siteAbs, nil, nil), nil
	}

	var raw monmsConfigFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("logging config: parse: %w", err)
	}
	return ConfigFromValues(siteAbs, raw.Logging, raw.LoggingRotation), nil
}

// ConfigFromValues builds logging config from config.json logging fields.
func ConfigFromValues(siteAbs string, levels []string, rot *RotationConfig) Config {
	dir := filepath.Join(siteAbs, ".monms", "logs")
	rotation := RotationConfig{
		MaxSizeMB:  10,
		MaxBackups: 5,
		MaxAgeDays: 30,
		Compress:   true,
	}
	if rot != nil {
		if rot.MaxSizeMB > 0 {
			rotation.MaxSizeMB = rot.MaxSizeMB
		}
		if rot.MaxBackups > 0 {
			rotation.MaxBackups = rot.MaxBackups
		}
		if rot.MaxAgeDays > 0 {
			rotation.MaxAgeDays = rot.MaxAgeDays
		}
		rotation.Compress = rot.Compress
	}
	if levels == nil {
		levels = DefaultLevelNames()
	}
	return Config{
		Dir:      dir,
		Flags:    ParseLevels(levels),
		Rotation: rotation,
	}
}

// ActiveFiles returns human-readable log filenames that will receive writes.
func (c Config) ActiveFiles() []string {
	out := []string{"pocketbase.log", "error.log"}
	if c.Flags.Has(FlagWarn) {
		out = append(out, "warn.log")
	}
	if c.Flags.Has(FlagInfo) {
		out = append(out, "info.log")
	}
	if c.Flags.Has(FlagDebug) {
		out = append(out, "debug.log")
	}
	if c.Flags.Has(FlagSchema) {
		out = append(out, "schema.log")
	}
	return out
}
