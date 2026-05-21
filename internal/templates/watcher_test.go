package templates

import (
	"context"
	"html/template"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatcherInvalidates(t *testing.T) {
	dir := t.TempDir()
	templatesDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	tplPath := filepath.Join(templatesDir, "page.gohtml")
	if err := os.WriteFile(tplPath, []byte(`{{define "body"}}v1{{end}}`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	cache := NewCache()
	cache.SetProductionMode(true)

	var loads int
	var mu sync.Mutex
	loader := func() (*template.Template, error) {
		mu.Lock()
		loads++
		mu.Unlock()
		return template.ParseFiles(tplPath)
	}

	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("prime Get: %v", err)
	}
	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("cached Get: %v", err)
	}
	mu.Lock()
	if loads != 1 {
		t.Fatalf("expected one load before watch, got %d", loads)
	}
	mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := StartWatcher(ctx, dir, cache.Flush); err != nil {
		t.Fatalf("StartWatcher: %v", err)
	}

	if err := os.WriteFile(tplPath, []byte(`{{define "body"}}v2{{end}}`), 0o644); err != nil {
		t.Fatalf("rewrite template: %v", err)
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if _, err := cache.Get("page", loader); err != nil {
			t.Fatalf("Get after change: %v", err)
		}
		mu.Lock()
		count := loads
		mu.Unlock()
		if count >= 2 {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	mu.Lock()
	count := loads
	mu.Unlock()
	t.Fatalf("expected cache flush within 200ms debounce window, loader count %d", count)
}

func TestWatcherIgnoresNonGohtml(t *testing.T) {
	dir := t.TempDir()
	templatesDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	tplPath := filepath.Join(templatesDir, "page.gohtml")
	if err := os.WriteFile(tplPath, []byte(`{{define "body"}}v1{{end}}`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	cache := NewCache()
	cache.SetProductionMode(true)

	var loads int
	loader := func() (*template.Template, error) {
		loads++
		return template.ParseFiles(tplPath)
	}

	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("prime Get: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := StartWatcher(ctx, dir, cache.Flush); err != nil {
		t.Fatalf("StartWatcher: %v", err)
	}

	cssPath := filepath.Join(dir, "assets", "main.css")
	if err := os.MkdirAll(filepath.Dir(cssPath), 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	if err := os.WriteFile(cssPath, []byte("body{}"), 0o644); err != nil {
		t.Fatalf("write css: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("Get after css write: %v", err)
	}
	if loads != 1 {
		t.Fatalf("expected non-.gohtml write to skip flush, loader count %d", loads)
	}
}

func TestWatcherAddsNewSubdirectory(t *testing.T) {
	dir := t.TempDir()
	templatesDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	cache := NewCache()
	cache.SetProductionMode(true)

	flushed := make(chan struct{}, 1)
	onChange := func() {
		cache.Flush()
		select {
		case flushed <- struct{}{}:
		default:
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := StartWatcher(ctx, dir, onChange); err != nil {
		t.Fatalf("StartWatcher: %v", err)
	}

	newDir := filepath.Join(templatesDir, "press")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("mkdir press: %v", err)
	}

	tplPath := filepath.Join(newDir, "index.gohtml")
	if err := os.WriteFile(tplPath, []byte(`{{define "body"}}press{{end}}`), 0o644); err != nil {
		t.Fatalf("write new template: %v", err)
	}

	select {
	case <-flushed:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected flush after creating template in new subdirectory")
	}
}
