package logging

import (
	"context"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/pocketbase/pocketbase/core"
)

var (
	globalCfg Config
	globalMu  sync.RWMutex
)

// Configure installs file logging for MonMS slog and returns the active config.
func Configure(siteAbs string) (Config, error) {
	cfg, err := LoadConfig(siteAbs)
	if err != nil {
		return Config{}, err
	}
	files, err := openFileSet(cfg)
	if err != nil {
		return Config{}, err
	}
	cfg.Files = files

	handler := newFileHandler(cfg.Flags, files)
	slog.SetDefault(slog.New(handler))

	globalMu.Lock()
	globalCfg = cfg
	globalMu.Unlock()
	return cfg, nil
}

// Dir returns the active log directory, or empty if not configured.
func Dir() string {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalCfg.Dir
}

// LogDirForSite returns the log directory path for a site without configuring writers.
func LogDirForSite(siteAbs string) string {
	return filepath.Join(siteAbs, ".monms", "logs")
}

// Schema writes a schema-level log record when schema logging is enabled.
func Schema(msg string, args ...any) {
	globalMu.RLock()
	cfg := globalCfg
	globalMu.RUnlock()
	if cfg.Files == nil || !cfg.Flags.Has(FlagSchema) {
		return
	}
	slog.Default().Log(context.Background(), LevelSchema, msg, args...)
}

// RegisterPocketBaseHook tees PocketBase logs into pocketbase.log (+ level files).
// Register after schema.RegisterBootstrapHook so the tee is installed before schema sync.
func RegisterPocketBaseHook(app core.App) {
	globalMu.RLock()
	cfg := globalCfg
	globalMu.RUnlock()
	if cfg.Files == nil {
		return
	}
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		disablePBSQLiteLogs(e.App)
		wrapPBLogger(e.App, cfg)
		return nil
	})
}
