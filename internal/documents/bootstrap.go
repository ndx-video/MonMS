package documents

import (
	"log/slog"
	"path/filepath"

	"github.com/monms/monms/internal/logging"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterBootstrapHook syncs markdown documents into PocketBase after schema bootstrap.
func RegisterBootstrapHook(app core.App, siteAbs string) {
	RegisterDoctreeRecordHooks(app, siteAbs)

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		result, err := SyncAll(e.App, siteAbs)
		if err != nil {
			return err
		}
		if result.Upserted > 0 {
			logging.Schema("documents sync: upserted markdown records",
				"collections", result.Collections,
				"upserted", result.Upserted,
			)
		}
		if len(result.Errors) > 0 {
			for _, msg := range result.Errors {
				slog.Warn("documents sync: file error", "error", msg)
			}
		}
		return nil
	})
}

// SchemaDir returns the schema directory for a site.
func SchemaDir(siteAbs string) string {
	return filepath.Join(siteAbs, "schema")
}
