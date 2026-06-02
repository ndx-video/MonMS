package mcp

import (
	"github.com/monms/monms/internal/content"
	"github.com/pocketbase/pocketbase/core"
)

// Deps configures the MCP HTTP server.
type Deps struct {
	App     core.App
	SiteAbs string
	Config  content.MonmsConfig
}
