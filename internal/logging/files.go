package logging

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// FileSet holds rotating log writers for a site.
type FileSet struct {
	PocketBase io.Writer
	Error      io.Writer
	Warn       io.Writer
	Info       io.Writer
	Debug      io.Writer
	Schema     io.Writer
}

func openFileSet(cfg Config) (*FileSet, error) {
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, err
	}
	mk := func(name string) io.Writer {
		return &lumberjack.Logger{
			Filename:   filepath.Join(cfg.Dir, name),
			MaxSize:    cfg.Rotation.MaxSizeMB,
			MaxBackups: cfg.Rotation.MaxBackups,
			MaxAge:     cfg.Rotation.MaxAgeDays,
			Compress:   cfg.Rotation.Compress,
		}
	}
	fs := &FileSet{
		PocketBase: mk("pocketbase.log"),
		Error:      mk("error.log"),
	}
	if cfg.Flags.Has(FlagWarn) {
		fs.Warn = mk("warn.log")
	}
	if cfg.Flags.Has(FlagInfo) {
		fs.Info = mk("info.log")
	}
	if cfg.Flags.Has(FlagDebug) {
		fs.Debug = mk("debug.log")
	}
	if cfg.Flags.Has(FlagSchema) {
		fs.Schema = mk("schema.log")
	}
	return fs, nil
}

func (fs *FileSet) writerForLevel(level slogLevelKind, flags LevelFlag) io.Writer {
	switch level {
	case kindSchema:
		if flags.Has(FlagSchema) && fs.Schema != nil {
			return fs.Schema
		}
	case kindError:
		return fs.Error
	case kindWarn:
		if flags.Has(FlagWarn) && fs.Warn != nil {
			return fs.Warn
		}
	case kindInfo:
		if flags.Has(FlagInfo) && fs.Info != nil {
			return fs.Info
		}
	case kindDebug:
		if flags.Has(FlagDebug) && fs.Debug != nil {
			return fs.Debug
		}
	}
	return nil
}

type slogLevelKind int

const (
	kindDebug slogLevelKind = iota
	kindInfo
	kindWarn
	kindError
	kindSchema
)

func classifyLevel(level int) slogLevelKind {
	switch {
	case level == int(LevelSchema):
		return kindSchema
	case level >= 8: // slog.LevelError
		return kindError
	case level >= 4: // slog.LevelWarn
		return kindWarn
	case level >= 0: // slog.LevelInfo
		return kindInfo
	default:
		return kindDebug
	}
}
