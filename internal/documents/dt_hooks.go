package documents

import (
	"os"
	"path/filepath"

	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterDoctreeRecordHooks syncs dt_* record changes back to markdown when not in dtFileSync.
func RegisterDoctreeRecordHooks(app core.App, siteAbs string) {
	app.OnRecordAfterCreateSuccess().BindFunc(recordHook(app, siteAbs))
	app.OnRecordAfterUpdateSuccess().BindFunc(recordHook(app, siteAbs))
	app.OnRecordAfterDeleteSuccess().BindFunc(deleteRecordHook(siteAbs))
}

func recordHook(app core.App, siteAbs string) func(*core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if DtSyncInProgress() {
			return nil
		}
		name := e.Record.Collection().Name
		if name == CollectionDtTrees || !schema.IsDoctreeCollection(name) {
			return nil
		}
		binding, err := FindBinding(siteAbs, name)
		if err != nil {
			return nil
		}
		return syncRecordFromHook(e.App, siteAbs, binding, e.Record)
	}
}

func deleteRecordHook(siteAbs string) func(*core.RecordEvent) error {
	return func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if DtSyncInProgress() {
			return nil
		}
		name := e.Record.Collection().Name
		if name == CollectionDtTrees || !schema.IsDoctreeCollection(name) {
			return nil
		}
		binding, err := FindBinding(siteAbs, name)
		if err != nil {
			return nil
		}
		rootAbs := filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root))
		pathKey := e.Record.GetString("path")
		if pathKey == "" {
			pathKey = e.Record.GetString("leaf_path")
		}
		if pathKey == "" {
			return nil
		}
		filePath := filepath.Join(rootAbs, filepath.FromSlash(pathKey)+".md")
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
}
