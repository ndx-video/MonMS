package templates

import "testing"

func TestResolveSlug(t *testing.T) {
	t.Skip("implemented in plan 01-03")

	// Table cases (D-10):
	// "/" -> templates/index.gohtml
	// "press" -> templates/press/index.gohtml
	// "press/2024" -> templates/press/2024.gohtml
}
