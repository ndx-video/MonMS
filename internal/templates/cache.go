package templates

import (
	"html/template"
	"sync"
)

// TemplateCache holds parsed templates in production mode only.
type TemplateCache struct {
	mu              sync.RWMutex
	cache           map[string]*template.Template
	productionMode  bool
}

// NewCache returns an empty template cache (development mode by default).
func NewCache() *TemplateCache {
	return &TemplateCache{
		cache: make(map[string]*template.Template),
	}
}

// SetProductionMode enables or disables in-memory caching.
func (c *TemplateCache) SetProductionMode(prod bool) {
	c.mu.Lock()
	c.productionMode = prod
	c.mu.Unlock()
}

// Active reports whether caching is enabled.
func (c *TemplateCache) Active() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.productionMode
}

// Get returns a cached template or loads it via loader when missing or inactive.
func (c *TemplateCache) Get(key string, loader func() (*template.Template, error)) (*template.Template, error) {
	if c.Active() {
		c.mu.RLock()
		if tmpl, ok := c.cache[key]; ok {
			c.mu.RUnlock()
			return tmpl, nil
		}
		c.mu.RUnlock()
	}

	tmpl, err := loader()
	if err != nil {
		return nil, err
	}

	if c.Active() {
		c.mu.Lock()
		if c.cache == nil {
			c.cache = make(map[string]*template.Template)
		}
		c.cache[key] = tmpl
		c.mu.Unlock()
	}

	return tmpl, nil
}

// Flush clears all cached templates.
func (c *TemplateCache) Flush() {
	c.mu.Lock()
	c.cache = make(map[string]*template.Template)
	c.mu.Unlock()
}
