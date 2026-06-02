package documents

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

const CollectionDtTrees = "dt_trees"

// LeafBindingSnapshot is one confirmed leaf collection binding.
type LeafBindingSnapshot struct {
	Collection string `json:"collection"`
	DestRoot   string `json:"destRoot"`
}

// DtTreeRow is the registry entry for one doctree stub.
type DtTreeRow struct {
	ID           string
	SourceRoot   string
	Status       string
	ConfirmedAt  string
	LeafBindings string
}

// EnsureDtTreesCollection creates the dt_trees registry collection if missing.
func EnsureDtTreesCollection(app core.App) error {
	if _, err := app.FindCollectionByNameOrId(CollectionDtTrees); err == nil {
		return nil
	}

	c := core.NewBaseCollection(CollectionDtTrees)
	c.Fields.Add(
		&core.TextField{Name: "id", Required: true, PrimaryKey: true},
		&core.TextField{Name: "source_root"},
		&core.SelectField{Name: "status", Values: []string{"pending", "active"}},
		&core.DateField{Name: "confirmed_at"},
		&core.JSONField{Name: "leaf_bindings"},
	)
	c.ListRule = nil
	c.ViewRule = nil
	c.CreateRule = nil
	c.UpdateRule = nil
	c.DeleteRule = nil
	return app.Save(c)
}

// UpsertDtTree writes or updates the registry row for a doctree stub.
func UpsertDtTree(app core.App, row DtTreeRow) error {
	if err := EnsureDtTreesCollection(app); err != nil {
		return err
	}

	coll, err := app.FindCollectionByNameOrId(CollectionDtTrees)
	if err != nil {
		return err
	}

	rec, err := app.FindRecordById(CollectionDtTrees, row.ID)
	if err != nil {
		rec = core.NewRecord(coll)
		rec.Set("id", row.ID)
	}

	rec.Set("source_root", row.SourceRoot)
	rec.Set("status", row.Status)
	if row.ConfirmedAt != "" {
		rec.Set("confirmed_at", row.ConfirmedAt)
	}
	if row.LeafBindings != "" {
		rec.Set("leaf_bindings", row.LeafBindings)
	}

	return app.Save(rec)
}

// BuildLeafBindingsJSON encodes binding candidates for dt_trees.leaf_bindings.
func BuildLeafBindingsJSON(candidates []BindingCandidate) (string, error) {
	snaps := make([]LeafBindingSnapshot, len(candidates))
	for i, c := range candidates {
		snaps[i] = LeafBindingSnapshot{Collection: c.Collection, DestRoot: c.DestRoot}
	}
	data, err := json.Marshal(snaps)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpsertDtTreeFromConfirm updates dt_trees after a successful migration confirm.
func UpsertDtTreeFromConfirm(app core.App, stub, sourceRoot string, candidates []BindingCandidate) error {
	bindingsJSON, err := BuildLeafBindingsJSON(candidates)
	if err != nil {
		return err
	}
	return UpsertDtTree(app, DtTreeRow{
		ID:           stub,
		SourceRoot:   sourceRoot,
		Status:       "active",
		ConfirmedAt:  time.Now().UTC().Format(time.RFC3339),
		LeafBindings: bindingsJSON,
	})
}

// SetDtTreePending marks a stub as pending migration.
func SetDtTreePending(app core.App, stub, sourceRoot string) error {
	return UpsertDtTree(app, DtTreeRow{
		ID:         stub,
		SourceRoot: sourceRoot,
		Status:     "pending",
	})
}

// DeleteDtTree removes a registry row (e.g. on prune).
func DeleteDtTree(app core.App, stub string) error {
	rec, err := app.FindRecordById(CollectionDtTrees, stub)
	if err != nil {
		return nil
	}
	return app.Delete(rec)
}

// ReadDtTree loads a registry row by stub id.
func ReadDtTree(app core.App, stub string) (DtTreeRow, bool, error) {
	rec, err := app.FindRecordById(CollectionDtTrees, stub)
	if err != nil {
		return DtTreeRow{}, false, nil
	}
	return DtTreeRow{
		ID:           rec.Id,
		SourceRoot:   rec.GetString("source_root"),
		Status:       rec.GetString("status"),
		ConfirmedAt:  rec.GetString("confirmed_at"),
		LeafBindings: fmt.Sprint(rec.Get("leaf_bindings")),
	}, true, nil
}
