package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/monms/monms/internal/apikeys"
	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/schema"
	"github.com/monms/monms/internal/validate"
)

func registerTools(s *server.MCPServer, deps Deps, pbBase string) {
	s.AddTool(mcp.NewTool("monms_list_collections",
		mcp.WithDescription("List editorial and markdown collection names from site schema"),
	), listCollectionsHandler(deps))

	s.AddTool(mcp.NewTool("monms_schema_list",
		mcp.WithDescription("List schema JSON filenames in site/schema"),
	), schemaListHandler(deps))

	s.AddTool(mcp.NewTool("monms_list_records",
		mcp.WithDescription("List records in an editorial PocketBase collection (respects key owner permissions)"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Collection name")),
	), listRecordsHandler(deps, pbBase))

	s.AddTool(mcp.NewTool("monms_get_record",
		mcp.WithDescription("Get one editorial collection record by id"),
		mcp.WithString("collection", mcp.Required()),
		mcp.WithString("id", mcp.Required()),
	), getRecordHandler(deps, pbBase))

	s.AddTool(mcp.NewTool("monms_update_record",
		mcp.WithDescription("Patch fields on an editorial collection record"),
		mcp.WithString("collection", mcp.Required()),
		mcp.WithString("id", mcp.Required()),
		mcp.WithObject("fields", mcp.Required(), mcp.Description("Fields to update")),
	), updateRecordHandler(deps, pbBase))

	s.AddTool(mcp.NewTool("monms_content_diff",
		mcp.WithDescription("Export pending editorial publish diff (publisher or superuser owners only)"),
	), contentDiffHandler(deps))

	s.AddTool(mcp.NewTool("monms_validate",
		mcp.WithDescription("Validate all site .gohtml templates and HTML balance"),
	), validateHandler(deps))

	registerDoctreeTools(s, deps)
}

func listCollectionsHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		editorial, err := schema.LoadEditorialCollectionNames(filepath.Join(deps.SiteAbs, "schema"))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		md, err := schema.LoadMarkdownBindings(filepath.Join(deps.SiteAbs, "schema"))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		var mdNames []string
		for _, b := range md {
			mdNames = append(mdNames, b.Name)
		}
		out := map[string]any{
			"editorial": editorial,
			"markdown":  mdNames,
		}
		raw, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func schemaListHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dir := filepath.Join(deps.SiteAbs, "schema")
		entries, err := os.ReadDir(dir)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		var names []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		raw, _ := json.Marshal(names)
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func listRecordsHandler(deps Deps, pbBase string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sess, err := requireSession(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		collection, err := req.RequireString("collection")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := assertPBNativeEditorial(deps.SiteAbs, collection); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		client := &PBClient{BaseURL: pbBase, Token: sess.OwnerToken}
		data, err := client.ListRecords(collection, 200)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func getRecordHandler(deps Deps, pbBase string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sess, err := requireSession(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		collection, err := req.RequireString("collection")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := assertPBNativeEditorial(deps.SiteAbs, collection); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		client := &PBClient{BaseURL: pbBase, Token: sess.OwnerToken}
		data, err := client.GetRecord(collection, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func updateRecordHandler(deps Deps, pbBase string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sess, err := requireSession(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		collection, err := req.RequireString("collection")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		args := req.GetArguments()
		fieldsRaw, ok := args["fields"]
		if !ok {
			return mcp.NewToolResultError("fields required"), nil
		}
		fields, ok := fieldsRaw.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("fields must be an object"), nil
		}
		if err := assertPBNativeEditorial(deps.SiteAbs, collection); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		client := &PBClient{BaseURL: pbBase, Token: sess.OwnerToken}
		data, err := client.PatchRecord(collection, id, fields)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func contentDiffHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sess, err := requireSession(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if !apikeys.IsSuperuserAuth(sess.Resolved.Owner) && !apikeys.IsPublisherOwner(sess.Resolved.Owner, deps.Config) {
			return mcp.NewToolResultError("publisher or superuser required for content diff"), nil
		}
		diff, err := content.DiffExport(deps.App, deps.SiteAbs)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		raw, _ := json.MarshalIndent(diff, "", "  ")
		return mcp.NewToolResultText(string(raw)), nil
	}
}

func validateHandler(deps Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := requireSession(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := validate.ValidateSiteTemplates(deps.SiteAbs); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText("validation passed"), nil
	}
}

func requireSession(ctx context.Context) (*session, error) {
	s, ok := sessionFromContext(ctx)
	if !ok || s == nil || s.Resolved == nil {
		return nil, fmt.Errorf("unauthorized")
	}
	return s, nil
}

func assertPBNativeEditorial(siteAbs, name string) error {
	names, err := content.LoadPBNativeEditorialCollectionNames(siteAbs)
	if err != nil {
		return err
	}
	for _, n := range names {
		if n == name {
			return nil
		}
	}
	return fmt.Errorf("collection %q is not a PocketBase-native editorial collection", name)
}
