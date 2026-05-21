package templates

import (
	"html/template"
	"testing"
)

func TestDevNoCache(t *testing.T) {
	cache := NewCache()
	cache.SetProductionMode(false)

	var loads int
	loader := func() (*template.Template, error) {
		loads++
		return template.New("page"), nil
	}

	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("first Get: %v", err)
	}
	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("second Get: %v", err)
	}

	if loads != 2 {
		t.Fatalf("expected loader invoked twice in dev mode, got %d", loads)
	}
}

func TestCacheFlush(t *testing.T) {
	cache := NewCache()
	cache.SetProductionMode(true)

	var loads int
	loader := func() (*template.Template, error) {
		loads++
		return template.New("page"), nil
	}

	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("first Get: %v", err)
	}
	if loads != 1 {
		t.Fatalf("expected one load after first Get, got %d", loads)
	}

	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("cached Get: %v", err)
	}
	if loads != 1 {
		t.Fatalf("expected cached hit without second load, got %d loads", loads)
	}

	cache.Flush()

	if _, err := cache.Get("page", loader); err != nil {
		t.Fatalf("Get after flush: %v", err)
	}
	if loads != 2 {
		t.Fatalf("expected loader invoked again after flush, got %d loads", loads)
	}
}

func TestActiveReflectsProductionMode(t *testing.T) {
	cache := NewCache()
	if cache.Active() {
		t.Fatal("expected inactive cache by default")
	}

	cache.SetProductionMode(true)
	if !cache.Active() {
		t.Fatal("expected active cache after SetProductionMode(true)")
	}
}
