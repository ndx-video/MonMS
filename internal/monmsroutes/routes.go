// Package monmsroutes defines canonical MonMS HTTP path prefixes.
//
// /api/monms/* — JSON REST only (machine clients).
// /_monms/*    — operator HTML tools and browser session helpers.
package monmsroutes

const (
	// DashboardHomePath is the operator dashboard entry point.
	DashboardHomePath = "/_monms/"

	// AdminPath is the PocketBase superuser admin SPA.
	AdminPath = "/_/"

	// ContentImportPath is the production editorial upsert API (Bearer MONMS_PUBLISH_TOKEN).
	ContentImportPath = "/api/monms/content/import"

	// PublishPath is the staging Publish to live HTML console.
	PublishPath = "/_monms/publish"
	// PublishDiffPath returns JSON diff data for the publish console.
	PublishDiffPath = "/_monms/publish/diff"

	// DocumentsPath is the markdown document browser in the operator dashboard.
	DocumentsPath = "/_monms/documents"

	// AuthSyncPath bridges PocketBase admin localStorage auth to the monms_auth cookie.
	AuthSyncPath = "/_monms/auth/sync"
	// AuthLogoutPath clears the monms_auth session cookie.
	AuthLogoutPath = "/_monms/auth/logout"

	// APIKeysPath is the operator API key management console.
	APIKeysPath = "/_monms/api-keys"
	// APIKeysRevokePath revokes an API key (POST).
	APIKeysRevokePath = "/_monms/api-keys/revoke"

	// MCPSettingsPath is the superuser MCP server settings console.
	MCPSettingsPath = "/_monms/mcp"

	// MCPHTTPPath is the Streamable HTTP endpoint path on the MCP listener (default for mcp-go).
	MCPHTTPPath = "/mcp"
)
