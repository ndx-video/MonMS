package content

import (
	"log/slog"

	"github.com/monms/monms/internal/site"
)

// ApplyShapeSyncFromSite runs optional shape sync from config before serve.
func ApplyShapeSyncFromSite(siteAbs string) error {
	cfg, err := LoadMonmsConfig(siteAbs)
	if err != nil {
		return err
	}

	opts, ok := site.ShapeSyncOptionsFromConfig(cfg.ShapeSync)
	if !ok {
		return nil
	}

	if err := site.Sync(siteAbs, opts); err != nil {
		if cfg.ShapeSync != nil && cfg.ShapeSync.FailOnError {
			return err
		}
		slog.Warn("site shape sync failed; continuing with current checkout", "err", err)
		return nil
	}

	return nil
}
