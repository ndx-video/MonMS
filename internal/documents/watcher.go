package documents

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

// StartWatcher watches documents/**/*.md and re-syncs into PocketBase after debounce.
func StartWatcher(ctx context.Context, siteAbs string, app core.App) error {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return err
	}
	if len(bindings) == 0 {
		return nil
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	syncFn := func() {
		result, syncErr := SyncAll(app, siteAbs)
		if syncErr != nil {
			slog.Warn("documents watcher sync failed", "error", syncErr)
			return
		}
		if result.Upserted > 0 {
			slog.Info("documents watcher synced", "upserted", result.Upserted)
		}
	}

	go func() {
		defer w.Close()

		var mu sync.Mutex
		var timer *time.Timer

		schedule := func() {
			mu.Lock()
			defer mu.Unlock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(200*time.Millisecond, syncFn)
		}

		for {
			select {
			case <-ctx.Done():
				mu.Lock()
				if timer != nil {
					timer.Stop()
				}
				mu.Unlock()
				return
			case e, ok := <-w.Events:
				if !ok {
					return
				}
				if e.Has(fsnotify.Create) {
					if fi, statErr := os.Stat(e.Name); statErr == nil && fi.IsDir() {
						_ = w.Add(e.Name)
					}
				}
				if !e.Has(fsnotify.Write) && !e.Has(fsnotify.Create) && !e.Has(fsnotify.Rename) {
					continue
				}
				if !strings.HasSuffix(strings.ToLower(e.Name), ".md") {
					continue
				}
				schedule()
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	for _, binding := range bindings {
		root := filepath.Join(siteAbs, filepath.FromSlash(binding.Monms.Root))
		if err := watchTree(w, root); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func watchTree(w *fsnotify.Watcher, root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return w.Add(path)
		}
		return nil
	})
}
