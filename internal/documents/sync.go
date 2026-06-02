package documents

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

// SyncResult summarizes one documents sync run.
type SyncResult struct {
	Collections int
	Upserted    int
	Skipped     int
	Errors      []string
}

// SyncAll upserts markdown-backed editorial collections into PocketBase.
func SyncAll(app core.App, siteAbs string) (SyncResult, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return SyncResult{}, err
	}
	if len(bindings) == 0 {
		return SyncResult{}, nil
	}

	var result SyncResult
	for _, binding := range bindings {
		n, errs := syncCollection(app, siteAbs, binding)
		result.Collections++
		result.Upserted += n
		result.Errors = append(result.Errors, errs...)
	}
	return result, nil
}

func syncCollection(app core.App, siteAbs string, binding schema.CollectionMeta) (int, []string) {
	root := filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root))
	docs, err := WalkMarkdownFiles(root)
	if err != nil {
		return 0, []string{fmt.Sprintf("%s: %v", binding.Name, err)}
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].FilePath < docs[j].FilePath
	})

	var upserted int
	var errs []string
	for _, doc := range docs {
		record, err := RecordFromDocument(binding, doc, root)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", doc.FilePath, err))
			continue
		}
		if err := upsertRecord(app, binding.Name, record); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", doc.FilePath, err))
			continue
		}
		upserted++
	}

	slog.Info("documents sync: collection complete",
		"collection", binding.Name,
		"files", len(docs),
		"upserted", upserted,
	)
	return upserted, errs
}

// DiffResult lists PB records in markdown collections with no backing file.
type DiffResult struct {
	Orphans []OrphanRecord
}

// OrphanRecord is a PocketBase record without a matching markdown file.
type OrphanRecord struct {
	Collection string
	ID         string
	Slug       string
}

// DiffOrphans compares markdown files to PB records for markdown collections.
func DiffOrphans(app core.App, siteAbs string) (DiffResult, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return DiffResult{}, err
	}

	var orphans []OrphanRecord
	for _, binding := range bindings {
		root := filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root))
		docs, err := WalkMarkdownFiles(root)
		if err != nil {
			return DiffResult{}, err
		}

		fileIDs := make(map[string]struct{}, len(docs))
		for _, doc := range docs {
			record, err := RecordFromDocument(binding, doc, root)
			if err != nil {
				continue
			}
			if id, _ := record["id"].(string); id != "" {
				fileIDs[id] = struct{}{}
			}
		}

		records, err := app.FindAllRecords(binding.Name)
		if err != nil {
			return DiffResult{}, fmt.Errorf("documents diff: list %s: %w", binding.Name, err)
		}
		for _, rec := range records {
			if _, ok := fileIDs[rec.Id]; !ok {
				orphans = append(orphans, OrphanRecord{
					Collection: binding.Name,
					ID:         rec.Id,
					Slug:       rec.GetString("slug"),
				})
			}
		}
	}

	sort.Slice(orphans, func(i, j int) bool {
		if orphans[i].Collection != orphans[j].Collection {
			return orphans[i].Collection < orphans[j].Collection
		}
		return orphans[i].ID < orphans[j].ID
	})

	return DiffResult{Orphans: orphans}, nil
}
