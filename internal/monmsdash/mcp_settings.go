package monmsdash

import (
	"strconv"
	"strings"

	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/monmsroutes"
	"github.com/pocketbase/pocketbase/core"
)

type mcpSettingsPageData struct {
	PageData
	MCP              content.MCPConfig
	RestartRequired  bool
	Saved            bool
}

func registerMCPSettingsRoutes(se *core.ServeEvent, deps Deps, tmpl *templates) {
	authBind := bindLoadAuth(deps.LoadAuth)
	authRedirect := requireAuthenticatedRedirect()
	super := requireSuperuser()

	se.Router.GET(monmsroutes.MCPSettingsPath, mcpSettingsGetHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(super)

	se.Router.POST(monmsroutes.MCPSettingsPath, mcpSettingsPostHandler(deps, tmpl)).
		BindFunc(authBind).
		BindFunc(authRedirect).
		BindFunc(super)
}

func mcpSettingsGetHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data, err := buildMCPSettingsPageData(e, deps, false, false)
		if err != nil {
			return e.InternalServerError("mcp settings", err)
		}
		return tmpl.renderPage(e.Response, "mcp", data)
	}
}

func mcpSettingsPostHandler(deps Deps, tmpl *templates) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if err := e.Request.ParseForm(); err != nil {
			return e.BadRequestError("invalid form", err)
		}

		mcpCfg := content.MCPConfig{
			Enabled:               e.Request.FormValue("enabled") == "on",
			Host:                  strings.TrimSpace(e.Request.FormValue("host")),
			Port:                  strings.TrimSpace(e.Request.FormValue("port")),
			AllowNonSuperuserKeys: e.Request.FormValue("allowNonSuperuserKeys") == "on",
		}
		if mcpCfg.Host == "" {
			mcpCfg.Host = content.DefaultMCPConfig().Host
		}
		if mcpCfg.Port == "" {
			mcpCfg.Port = content.DefaultMCPConfig().Port
		}
		if _, err := strconv.Atoi(mcpCfg.Port); err != nil {
			data, derr := buildMCPSettingsPageData(e, deps, false, false)
			if derr != nil {
				return e.InternalServerError("mcp settings", derr)
			}
			data.FlashMessage = "Port must be a number."
			data.FlashError = true
			return tmpl.renderPage(e.Response, "mcp", data)
		}

		if err := content.SaveMonmsMCPSettings(deps.SiteAbs, mcpCfg); err != nil {
			return e.InternalServerError("save mcp settings", err)
		}

		data, err := buildMCPSettingsPageData(e, deps, true, true)
		if err != nil {
			return e.InternalServerError("mcp settings", err)
		}
		data.FlashMessage = "MCP settings saved. Restart monms for bind changes to take effect."
		return tmpl.renderPage(e.Response, "mcp", data)
	}
}

func buildMCPSettingsPageData(e *core.RequestEvent, deps Deps, saved, restartRequired bool) (mcpSettingsPageData, error) {
	cfg, err := content.LoadMonmsConfig(deps.SiteAbs)
	if err != nil {
		return mcpSettingsPageData{}, err
	}
	base, err := buildPageData(e, deps.SiteAbs, "mcp", "MCP Settings")
	if err != nil {
		return mcpSettingsPageData{}, err
	}
	return mcpSettingsPageData{
		PageData:        base,
		MCP:             cfg.MCP,
		Saved:           saved,
		RestartRequired: restartRequired,
	}, nil
}
