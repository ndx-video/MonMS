package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/monms/monms/internal/documents"
	"github.com/monms/monms/internal/schema"
)

func registerDoctreeTools(s *server.MCPServer, deps Deps) {
	s.AddTool(mcp.NewTool("monms_doctree_bindings",
		mcp.WithDescription("List markdown collection bindings from site schema (monms.source=markdown)"),
	), doctreeBindingsHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_forest",
		mcp.WithDescription("Build nested document trees from synced PocketBase index (folders, slugs, orphan count)"),
	), doctreeForestHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_list",
		mcp.WithDescription("List markdown files under a bound collection root (Git-canonical filesystem)"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Markdown collection name")),
		mcp.WithString("folder", mcp.Description("Optional path prefix filter (e.g. guides)")),
	), doctreeListHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_get",
		mcp.WithDescription("Read one markdown document by public slug or by collection+path"),
		mcp.WithString("slug", mcp.Description("Public URL slug (e.g. guides/setup); searches all markdown collections")),
		mcp.WithString("collection", mcp.Description("Collection name when using path lookup")),
		mcp.WithString("path", mcp.Description("Path key without .md (e.g. guides/setup)")),
	), doctreeGetHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_write",
		mcp.WithDescription("Create or update a markdown file under a bound collection; optionally sync to PocketBase"),
		mcp.WithString("collection", mcp.Required()),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path key without .md")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Markdown body (after frontmatter)")),
		mcp.WithObject("meta", mcp.Description("YAML frontmatter fields (id, title, date, …)")),
		mcp.WithBoolean("sync", mcp.Description("Upsert into PocketBase after write (default true)")),
	), doctreeWriteHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_delete",
		mcp.WithDescription("Delete a markdown file under a bound collection; optionally sync PocketBase index"),
		mcp.WithString("collection", mcp.Required()),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path key without .md")),
		mcp.WithBoolean("sync", mcp.Description("Run documents sync after delete (default true)")),
	), doctreeDeleteHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_sync",
		mcp.WithDescription("Sync markdown files from Git tree into PocketBase derived index"),
		mcp.WithString("collection", mcp.Description("Optional: sync one collection only")),
	), doctreeSyncHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_diff",
		mcp.WithDescription("List PocketBase markdown records with no backing .md file (orphans)"),
	), doctreeDiffHandler(deps))

	s.AddTool(mcp.NewTool("monms_doctree_sections",
		mcp.WithDescription("Parse heading-bounded sections from a document (for template docSection/docHeading)"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Public URL slug")),
		mcp.WithNumber("level", mcp.Description("Optional: filter to heading level 1–6")),
		mcp.WithNumber("index", mcp.Description("Optional: filter to per-level index (requires level)")),
	), doctreeSectionsHandler(deps))
}

func doctreeBindingsHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		bindings, err := schema.LoadMarkdownBindings(documents.SchemaDir(deps.SiteAbs))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		type bindingView struct {
			Collection string            `json:"collection"`
			Root       string            `json:"root"`
			SlugFrom   string            `json:"slugFrom,omitempty"`
			IDFrom     string            `json:"idFrom,omitempty"`
			BodyField  string            `json:"bodyField,omitempty"`
			Fields     map[string]string `json:"fields,omitempty"`
		}
		out := make([]bindingView, 0, len(bindings))
		for _, b := range bindings {
			body := b.Monms.Body
			if body == "" {
				body = "body"
			}
			out = append(out, bindingView{
				Collection: b.Name,
				Root:       b.Monms.Root,
				SlugFrom:   b.Monms.SlugFrom,
				IDFrom:     b.Monms.IDFrom,
				BodyField:  body,
				Fields:     b.Monms.Fields,
			})
		}
		raw, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeForestHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		forest, err := documents.BuildForest(deps.App, deps.SiteAbs)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		raw, _ := json.MarshalIndent(forest, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeListHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		collection, err := req.RequireString("collection")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		binding, err := documents.FindBinding(deps.SiteAbs, collection)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		entries, err := documents.ListDocuments(deps.SiteAbs, binding)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		folder := strings.Trim(strings.TrimSpace(req.GetString("folder", "")), "/")
		if folder != "" {
			prefix := folder + "/"
			filtered := entries[:0]
			for _, e := range entries {
				if e.PathKey == folder || strings.HasPrefix(e.PathKey, prefix) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
		raw, _ := json.MarshalIndent(entries, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeGetHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		slug := strings.Trim(req.GetString("slug", ""), "/")
		collection := req.GetString("collection", "")
		pathKey := strings.Trim(req.GetString("path", ""), "/")

		if slug != "" {
			match, err := documents.FindBySlug(deps.App, deps.SiteAbs, slug)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if match == nil {
				return mcp.NewToolResultError(fmt.Sprintf("document not found: %s", slug)), nil
			}
			bodyField := match.Binding.Monms.Body
			if bodyField == "" {
				bodyField = "body"
			}
			filePath := documents.DocFilePath(deps.SiteAbs, match.Binding, match.Record.GetString("path"))
			out := map[string]any{
				"collection": match.Collection,
				"slug":       match.Record.GetString("slug"),
				"path":       match.Record.GetString("path"),
				"id":         match.Record.Id,
				"filePath":   filePath,
				"body":       match.Record.GetString(bodyField),
			}
			if _, err := os.Stat(filePath); err == nil {
				doc, err := documents.ParseFile(filePath)
				if err == nil {
					out["meta"] = doc.Meta
					out["body"] = doc.Body
					out["source"] = "filesystem"
				}
			} else {
				out["source"] = "pocketbase"
			}
			raw, _ := json.MarshalIndent(out, "", "  ")
			return mcp.NewToolResultText(string(raw)), nil
		}

		if collection == "" || pathKey == "" {
			return mcp.NewToolResultError("provide slug or both collection and path"), nil
		}
		binding, doc, err := documents.ReadDocument(deps.SiteAbs, collection, pathKey)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		record, err := documents.RecordFromDocument(binding, doc, filepath.Join(deps.SiteAbs, filepath.FromSlash(binding.Monms.Root)))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out := map[string]any{
			"collection": collection,
			"slug":       record["slug"],
			"path":       record["path"],
			"id":         record["id"],
			"filePath":   documents.DocFilePath(deps.SiteAbs, binding, pathKey),
			"meta":       doc.Meta,
			"body":       doc.Body,
			"source":     "filesystem",
		}
		raw, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeWriteHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		collection, err := req.RequireString("collection")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		pathKey, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body, err := req.RequireString("body")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		pathKey = strings.Trim(pathKey, "/")

		binding, err := documents.FindBinding(deps.SiteAbs, collection)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		meta := map[string]any{}
		if raw, ok := req.GetArguments()["meta"]; ok && raw != nil {
			m, ok := raw.(map[string]any)
			if !ok {
				return mcp.NewToolResultError("meta must be an object"), nil
			}
			for k, v := range m {
				meta[k] = v
			}
		}
		if _, ok := meta["id"]; !ok {
			meta["id"] = documents.StableRecordID(collection, pathKey)
		}
		if _, ok := meta["title"]; !ok {
			base := filepath.Base(pathKey)
			meta["title"] = strings.ReplaceAll(base, "-", " ")
		}

		filePath := documents.DocFilePath(deps.SiteAbs, binding, pathKey)
		if err := documents.WriteFile(filePath, meta, body); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		sync := true
		if v, ok := req.GetArguments()["sync"]; ok {
			if b, ok := v.(bool); ok {
				sync = b
			}
		}

		result := map[string]any{
			"written":  filePath,
			"collection": collection,
			"path":     pathKey,
			"id":       meta["id"],
		}
		if sync {
			n, syncErr := documents.SyncCollection(deps.App, deps.SiteAbs, collection)
			result["synced"] = n
			if syncErr != nil {
				result["syncError"] = syncErr.Error()
			}
		}
		raw, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeDeleteHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		collection, err := req.RequireString("collection")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		pathKey, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		pathKey = strings.Trim(pathKey, "/")

		binding, err := documents.FindBinding(deps.SiteAbs, collection)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		filePath := documents.DocFilePath(deps.SiteAbs, binding, pathKey)
		if err := os.Remove(filePath); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		sync := true
		if v, ok := req.GetArguments()["sync"]; ok {
			if b, ok := v.(bool); ok {
				sync = b
			}
		}
		result := map[string]any{"deleted": filePath}
		if sync {
			n, syncErr := documents.SyncCollection(deps.App, deps.SiteAbs, collection)
			result["synced"] = n
			if syncErr != nil {
				result["syncError"] = syncErr.Error()
			}
		}
		raw, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeSyncHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		collection := strings.TrimSpace(req.GetString("collection", ""))
		if collection != "" {
			n, err := documents.SyncCollection(deps.App, deps.SiteAbs, collection)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			raw, _ := json.MarshalIndent(map[string]any{"collection": collection, "upserted": n}, "", "  ")
			return mcp.NewToolResultText(string(raw)), nil
		}
		result, err := documents.SyncAll(deps.App, deps.SiteAbs)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		raw, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeDiffHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		diff, err := documents.DiffOrphans(deps.App, deps.SiteAbs)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		raw, _ := json.MarshalIndent(diff, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func doctreeSectionsHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		match, err := documents.FindBySlug(deps.App, deps.SiteAbs, slug)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if match == nil {
			return mcp.NewToolResultError(fmt.Sprintf("document not found: %s", slug)), nil
		}
		bodyField := match.Binding.Monms.Body
		if bodyField == "" {
			bodyField = "body"
		}
		body := match.Record.GetString(bodyField)
		sections, err := documents.ParseSections(body)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		args := req.GetArguments()
		if levelRaw, ok := args["level"]; ok {
			level, ok := levelRaw.(float64)
			if !ok {
				return mcp.NewToolResultError("level must be a number"), nil
			}
			if indexRaw, hasIndex := args["index"]; hasIndex {
				index, ok := indexRaw.(float64)
				if !ok {
					return mcp.NewToolResultError("index must be a number"), nil
				}
				sec, found := documents.SectionAt(sections, int(level), int(index))
				if !found {
					return mcp.NewToolResultError(fmt.Sprintf("no section at level %d index %d", int(level), int(index))), nil
				}
				raw, _ := json.MarshalIndent(sec, "", "  ")
				return mcp.NewToolResultText(string(raw)), nil
			}
			filtered := sections[:0]
			for _, s := range sections {
				if s.Level == int(level) {
					filtered = append(filtered, s)
				}
			}
			sections = filtered
		}
		raw, _ := json.MarshalIndent(sections, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}
