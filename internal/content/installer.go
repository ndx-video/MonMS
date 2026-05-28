package content

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// WrapInstallerFunc replaces PocketBase's bind-address installer URL with the
// same public origin used by the MonMS startup banner (siteUrl or allowedHosts).
func WrapInstallerFunc(
	inner func(app core.App, systemSuperuser *core.Record, baseURL string) error,
	siteAbs string,
	serveArgs []string,
) func(app core.App, systemSuperuser *core.Record, baseURL string) error {
	if inner == nil {
		inner = apis.DefaultInstallerFunc
	}
	return func(app core.App, systemSuperuser *core.Record, _ string) error {
		base := installerBaseURL(siteAbs, serveArgs)
		return inner(app, systemSuperuser, base)
	}
}

func installerBaseURL(siteAbs string, serveArgs []string) string {
	cfg, err := LoadMonmsConfig(siteAbs)
	if err != nil {
		return resolveDisplayBase(MonmsConfig{}, serveArgs)
	}
	return resolveDisplayBase(cfg, serveArgs)
}
