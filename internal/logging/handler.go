package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// fileHandler routes slog records to rotating level files.
type fileHandler struct {
	flags LevelFlag
	files *FileSet
	attrs []slog.Attr
	group string
	mu    sync.Mutex
}

func newFileHandler(flags LevelFlag, files *FileSet) *fileHandler {
	return &fileHandler{flags: flags, files: files}
}

func (h *fileHandler) Enabled(_ context.Context, level slog.Level) bool {
	flag := slogLevelFlag(level)
	if level == LevelSchema {
		return h.flags.Has(FlagSchema)
	}
	if flag == FlagError {
		return true
	}
	return h.flags.Has(flag)
}

func (h *fileHandler) Handle(_ context.Context, r slog.Record) error {
	kind := classifyLevel(int(r.Level))
	w := h.files.writerForLevel(kind, h.flags)
	if w == nil && r.Level != LevelSchema {
		if kind == kindError {
			w = h.files.Error
		}
	}
	if w == nil {
		return nil
	}
	line := formatRecord(r, h.group, h.attrs)
	_, err := io.WriteString(w, line)
	return err
}

func (h *fileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &h2
}

func (h *fileHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	h2.group = name
	return &h2
}

func formatRecord(r slog.Record, group string, baseAttrs []slog.Attr) string {
	var b strings.Builder
	b.WriteString(r.Time.Format(time.RFC3339))
	b.WriteByte(' ')
	b.WriteString(levelName(r.Level))
	if group != "" {
		b.WriteByte(' ')
		b.WriteString(group)
		b.WriteByte('.')
	}
	b.WriteByte(' ')
	b.WriteString(r.Message)

	attrs := append([]slog.Attr{}, baseAttrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	for _, a := range attrs {
		b.WriteByte(' ')
		b.WriteString(a.Key)
		b.WriteByte('=')
		b.WriteString(fmt.Sprint(a.Value.Any()))
	}
	b.WriteByte('\n')
	return b.String()
}

