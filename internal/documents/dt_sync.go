package documents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

var dtSyncInProgress atomic.Bool

// DtFileSyncResult summarizes a three-phase dt_* sync run.
type DtFileSyncResult struct {
	Collection     string
	RecordsCreated int
	RecordsUpdated int
	FilesCreated   int
	FilesUpdated   int
	Warnings       []string
}

// DtFileSyncAll runs DtFileSyncCollection for every dt_* markdown binding.
func DtFileSyncAll(app core.App, siteAbs string) ([]DtFileSyncResult, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return nil, err
	}
	var results []DtFileSyncResult
	for _, b := range bindings {
		if !schema.IsDoctreeCollection(b.Name) {
			continue
		}
		r, err := DtFileSyncCollection(app, siteAbs, b)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}
	return results, nil
}

// DtFileSyncCollection enforces FS↔PB existence, structural fields, and content LWW for one dt_* binding.
func DtFileSyncCollection(app core.App, siteAbs string, binding schema.CollectionMeta) (DtFileSyncResult, error) {
	if !binding.IsMarkdownSource() || !schema.IsDoctreeCollection(binding.Name) {
		return DtFileSyncResult{}, fmt.Errorf("documents: %q is not a dt_* markdown collection", binding.Name)
	}

	dtSyncInProgress.Store(true)
	defer dtSyncInProgress.Store(false)

	result := DtFileSyncResult{Collection: binding.Name}
	rootAbs := filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root))

	docs, err := WalkMarkdownFiles(rootAbs)
	if err != nil {
		return result, err
	}

	fileByID := make(map[string]ParsedDocument)
	for _, doc := range docs {
		record, err := RecordFromDocument(binding, doc, rootAbs)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		id, _ := record["id"].(string)
		if id != "" {
			fileByID[id] = doc
		}
	}

	records, err := app.FindAllRecords(binding.Name)
	if err != nil {
		return result, fmt.Errorf("dt sync list %s: %w", binding.Name, err)
	}

	recordByID := make(map[string]*core.Record, len(records))
	for _, rec := range records {
		recordByID[rec.Id] = rec
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Phase 1–3: filesystem → PocketBase
	for id, doc := range fileByID {
		record, err := RecordFromDocument(binding, doc, rootAbs)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		record["monms_sync_at"] = now

		rec, has := recordByID[id]
		if !has {
			if err := upsertRecord(app, binding.Name, record); err != nil {
				result.Warnings = append(result.Warnings, err.Error())
				continue
			}
			result.RecordsCreated++
			continue
		}

		if err := applyContentLWW(binding, doc, rec, record, &result); err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		if err := upsertRecord(app, binding.Name, record); err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		result.RecordsUpdated++
	}

	// Phase 1: PocketBase → filesystem (missing files)
	for id, rec := range recordByID {
		if _, ok := fileByID[id]; ok {
			continue
		}
		created, err := writeRecordToMarkdown(binding, rootAbs, rec)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		if created {
			result.FilesCreated++
		}
	}

	return result, nil
}

func applyContentLWW(binding schema.CollectionMeta, doc ParsedDocument, rec *core.Record, record map[string]any, result *DtFileSyncResult) error {
	bodyField := binding.Monms.Body
	if bodyField == "" {
		bodyField = defaultBodyField
	}

	fileTime := fileModTime(doc.FilePath)
	recTime := recordModTime(rec)

	recTitle := rec.GetString("title")
	recBody := rec.GetString(bodyField)

	if fileTime.After(recTime) {
		record["ts_mod"] = fileTime.UTC().Format(time.RFC3339)
		return nil
	}

	record["title"] = recTitle
	record[bodyField] = recBody
	record["ts_mod"] = rec.GetString("ts_mod")

	if recTitle == doc.Meta["title"] && recBody == doc.Body {
		return nil
	}

	meta := map[string]any{}
	for k, v := range doc.Meta {
		meta[k] = v
	}
	meta["title"] = recTitle
	if id, ok := record["id"].(string); ok {
		meta["id"] = id
	}
	if ts := rec.GetString("ts_mod"); ts != "" {
		meta["ts_mod"] = ts
	}
	if err := WriteFile(doc.FilePath, meta, recBody); err != nil {
		return err
	}
	result.FilesUpdated++
	return nil
}

func writeRecordToMarkdown(binding schema.CollectionMeta, rootAbs string, rec *core.Record) (bool, error) {
	pathKey := rec.GetString("path")
	if pathKey == "" {
		pathKey = rec.GetString("leaf_path")
	}
	if pathKey == "" {
		return false, fmt.Errorf("documents: record %s missing path", rec.Id)
	}

	filePath := filepath.Join(rootAbs, filepath.FromSlash(pathKey)+".md")
	if _, err := os.Stat(filePath); err == nil {
		return false, nil
	}

	bodyField := binding.Monms.Body
	if bodyField == "" {
		bodyField = defaultBodyField
	}

	meta := map[string]any{
		"id":    rec.Id,
		"title": rec.GetString("title"),
	}
	if d := rec.GetString("doctree_id"); d != "" {
		meta["doctree_id"] = d
	}
	if lp := rec.GetString("leaf_path"); lp != "" {
		meta["leaf_path"] = lp
	}
	if ts := rec.GetString("ts_mod"); ts != "" {
		meta["ts_mod"] = ts
	}
	if desc := rec.GetString("description"); desc != "" {
		meta["description"] = desc
	}
	if err := WriteFile(filePath, meta, rec.GetString(bodyField)); err != nil {
		return false, err
	}
	return true, nil
}

func fileModTime(path string) time.Time {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}

// DtSyncInProgress reports whether dtFileSync is writing records or files.
func DtSyncInProgress() bool {
	return dtSyncInProgress.Load()
}

func recordModTime(rec *core.Record) time.Time {
	if rec == nil {
		return time.Time{}
	}
	if s := strings.TrimSpace(rec.GetString("ts_mod")); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
	}
	return rec.GetDateTime("updated").Time()
}

// syncRecordFromHook applies content LWW for a single dt_* record change from PocketBase.
func syncRecordFromHook(app core.App, siteAbs string, binding schema.CollectionMeta, rec *core.Record) error {
	dtSyncInProgress.Store(true)
	defer dtSyncInProgress.Store(false)

	rootAbs := filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root))
	pathKey := rec.GetString("path")
	if pathKey == "" {
		pathKey = rec.GetString("leaf_path")
	}
	if pathKey == "" {
		return fmt.Errorf("documents: record %s missing path", rec.Id)
	}
	filePath := filepath.Join(rootAbs, filepath.FromSlash(pathKey)+".md")

	bodyField := binding.Monms.Body
	if bodyField == "" {
		bodyField = defaultBodyField
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = writeRecordToMarkdown(binding, rootAbs, rec)
			return err
		}
		return err
	}

	recTime := recordModTime(rec)
	if fi.ModTime().After(recTime) {
		doc, err := ParseFile(filePath)
		if err != nil {
			return err
		}
		record, err := RecordFromDocument(binding, doc, rootAbs)
		if err != nil {
			return err
		}
		record["monms_sync_at"] = time.Now().UTC().Format(time.RFC3339)
		record["ts_mod"] = fi.ModTime().UTC().Format(time.RFC3339)
		return upsertRecord(app, binding.Name, record)
	}

	meta := map[string]any{"id": rec.Id, "title": rec.GetString("title")}
	if ts := rec.GetString("ts_mod"); ts != "" {
		meta["ts_mod"] = ts
	}
	if d := rec.GetString("doctree_id"); d != "" {
		meta["doctree_id"] = d
	}
	if lp := rec.GetString("leaf_path"); lp != "" {
		meta["leaf_path"] = lp
	}
	return WriteFile(filePath, meta, rec.GetString(bodyField))
}

// SyncBinding chooses DtFileSyncCollection for dt_* collections else legacy syncCollection.
func SyncBinding(app core.App, siteAbs string, binding schema.CollectionMeta) (int, []string, error) {
	if schema.IsDoctreeCollection(binding.Name) {
		r, err := DtFileSyncCollection(app, siteAbs, binding)
		if err != nil {
			return 0, r.Warnings, err
		}
		n := r.RecordsCreated + r.RecordsUpdated
		return n, r.Warnings, nil
	}
	n, errs := syncCollection(app, siteAbs, binding)
	return n, errs, nil
}

// SyncAfterBindDt runs DtFileSync for all dt_* bindings and returns per-collection results.
func SyncAfterBindDt(app core.App, siteAbs string) ([]DtFileSyncResult, error) {
	return DtFileSyncAll(app, siteAbs)
}
