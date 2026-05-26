package content

import (
	"fmt"
	"sort"

	"github.com/pocketbase/pocketbase/core"
)

// DiffResult summarizes pending editorial changes (PUB-09).
type DiffResult struct {
	HasChanges bool
	Changes    []string
	Checksum   string
}

// DiffExport compares live editorial export to publish-state checksum (PUB-09).
func DiffExport(app core.App, wsAbs string) (DiffResult, error) {
	current, err := ExportSnapshot(app, wsAbs)
	if err != nil {
		return DiffResult{}, err
	}

	checksum, err := ChecksumExport(current)
	if err != nil {
		return DiffResult{}, err
	}

	state, err := ReadPublishState(wsAbs)
	if err != nil {
		return DiffResult{}, err
	}

	result := DiffResult{Checksum: checksum}
	if state.Checksum == "" || checksum != state.Checksum {
		result.HasChanges = true
		baseline, err := LoadContentFiles(wsAbs)
		if err != nil {
			return DiffResult{}, err
		}
		result.Changes = diffSnapshots(baseline, current)
		if len(result.Changes) == 0 && state.Checksum != "" {
			result.Changes = []string{"editorial content changed since last publish"}
		}
	}
	return result, nil
}

func diffSnapshots(baseline, current []CollectionFile) []string {
	baseByColl := make(map[string]map[string]map[string]any)
	for _, f := range baseline {
		byID := make(map[string]map[string]any)
		for _, rec := range f.Records {
			id, _ := rec["id"].(string)
			if id != "" {
				byID[id] = rec
			}
		}
		baseByColl[f.Collection] = byID
	}

	curByColl := make(map[string]map[string]map[string]any)
	for _, f := range current {
		byID := make(map[string]map[string]any)
		for _, rec := range f.Records {
			id, _ := rec["id"].(string)
			if id != "" {
				byID[id] = rec
			}
		}
		curByColl[f.Collection] = byID
	}

	var changes []string
	for _, f := range current {
		baseRecords := baseByColl[f.Collection]
		for _, rec := range f.Records {
			id, _ := rec["id"].(string)
			if id == "" {
				continue
			}
			oldRec := baseRecords[id]
			if oldRec == nil {
				changes = append(changes, fmt.Sprintf("%s/%s: new record", f.Collection, id))
				continue
			}
			changes = append(changes, diffRecordFields(f.Collection, id, oldRec, rec)...)
		}
	}

	for coll, baseRecords := range baseByColl {
		curRecords := curByColl[coll]
		for id := range baseRecords {
			if curRecords == nil || curRecords[id] == nil {
				changes = append(changes, fmt.Sprintf("%s/%s: record deleted", coll, id))
			}
		}
	}

	sort.Strings(changes)
	return changes
}

func diffRecordFields(collection, id string, oldRec, newRec map[string]any) []string {
	keys := make(map[string]struct{})
	for k := range oldRec {
		if k != "id" {
			keys[k] = struct{}{}
		}
	}
	for k := range newRec {
		if k != "id" {
			keys[k] = struct{}{}
		}
	}

	var names []string
	for k := range keys {
		names = append(names, k)
	}
	sort.Strings(names)

	var changes []string
	for _, k := range names {
		oldV := fmt.Sprint(oldRec[k])
		newV := fmt.Sprint(newRec[k])
		if oldV != newV {
			changes = append(changes,
				fmt.Sprintf("%s/%s/%s: %q → %q", collection, id, k, oldV, newV))
		}
	}
	return changes
}
