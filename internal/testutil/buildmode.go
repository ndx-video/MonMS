package testutil

import "testing"

// WithProduction runs fn in a production-mode test context.
//
// Production cache and watcher behavior is controlled by compile-time ldflags on
// main.buildMode, e.g.:
//
//	go test -ldflags "-X main.buildMode=production" ./internal/templates/...
//
// Plans that need production mode should prefer ldflags over t.Setenv.
func WithProduction(t *testing.T, fn func()) {
	t.Helper()
	t.Skip("production mode tests require -ldflags \"-X main.buildMode=production\"; implemented in plan 01-02")
}
