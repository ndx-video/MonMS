package logging

import (
	"context"
	"log/slog"
	"reflect"
	"unsafe"

	"github.com/pocketbase/pocketbase/core"
)

type pbTeeHandler struct {
	inner slog.Handler
	cfg   Config
}

func (h *pbTeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *pbTeeHandler) Handle(ctx context.Context, r slog.Record) error {
	writeSlogRecord(h.cfg, r)
	return h.inner.Handle(ctx, r)
}

func (h *pbTeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &pbTeeHandler{inner: h.inner.WithAttrs(attrs), cfg: h.cfg}
}

func (h *pbTeeHandler) WithGroup(name string) slog.Handler {
	return &pbTeeHandler{inner: h.inner.WithGroup(name), cfg: h.cfg}
}

func writeSlogRecord(cfg Config, r slog.Record) {
	if cfg.Files == nil {
		return
	}
	line := formatRecord(r, "", nil)
	_, _ = cfg.Files.PocketBase.Write([]byte(line))

	kind := classifyLevel(int(r.Level))
	w := cfg.Files.writerForLevel(kind, cfg.Flags)
	if w == nil && kind == kindError {
		w = cfg.Files.Error
	}
	if w != nil {
		_, _ = w.Write([]byte(line))
	}
}

func wrapPBLogger(app core.App, cfg Config) {
	base, ok := app.(*core.BaseApp)
	if !ok || base.Logger() == nil {
		return
	}
	inner := base.Logger().Handler()
	setBaseAppLogger(base, slog.New(&pbTeeHandler{inner: inner, cfg: cfg}))
}

func disablePBSQLiteLogs(app core.App) {
	settings := app.Settings()
	if settings.Logs.MaxDays == 0 {
		return
	}
	settings.Logs.MaxDays = 0
	_ = app.Save(settings)
}

func setBaseAppLogger(app *core.BaseApp, logger *slog.Logger) {
	field := reflect.ValueOf(app).Elem().FieldByName("logger")
	ptr := unsafe.Pointer(field.UnsafeAddr())
	reflect.NewAt(field.Type(), ptr).Elem().Set(reflect.ValueOf(logger))
}
