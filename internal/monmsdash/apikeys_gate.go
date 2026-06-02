package monmsdash

import (
	"github.com/monms/monms/internal/apikeys"
	"github.com/monms/monms/internal/content"
	"github.com/pocketbase/pocketbase/core"
)

func requireCanManageAPIKeys(siteAbs string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		cfg, err := content.LoadMonmsConfig(siteAbs)
		if err != nil {
			return e.InternalServerError("config", err)
		}
		if !apikeys.CanManageKeys(e.Auth, cfg) {
			return e.ForbiddenError("api key access not allowed", nil)
		}
		return e.Next()
	}
}
