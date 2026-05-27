package templates

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// StartWatcher recursively watches siteRoot for .gohtml changes and calls onChange after debounce.
func StartWatcher(ctx context.Context, siteRoot string, onChange func()) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer w.Close()

		var mu sync.Mutex
		timers := map[string]*time.Timer{}

		flush := func() {
			if onChange != nil {
				onChange()
			}
		}

		for {
			select {
			case <-ctx.Done():
				mu.Lock()
				for _, timer := range timers {
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
				if !e.Has(fsnotify.Write) && !e.Has(fsnotify.Create) {
					continue
				}
				if !strings.HasSuffix(e.Name, ".gohtml") {
					continue
				}

				mu.Lock()
				if t, ok := timers[e.Name]; ok {
					t.Reset(100 * time.Millisecond)
				} else {
					path := e.Name
					timers[path] = time.AfterFunc(100*time.Millisecond, func() {
						flush()
						mu.Lock()
						delete(timers, path)
						mu.Unlock()
					})
				}
				mu.Unlock()
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return filepath.WalkDir(siteRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return w.Add(path)
		}
		return nil
	})
}
