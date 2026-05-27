package router

import (
	_ "embed"
	"testing/fstest"

	"github.com/pocketbase/pocketbase/core"
)

//go:embed embed/admin/main.js
var adminMainJS []byte

// RegisterAdminUIExtension adds a PocketBase superuser UI extension that links
// to the public site frontend from the admin header (PocketBase v0.38+ UIExtensions).
func RegisterAdminUIExtension(se *core.ServeEvent) {
	se.UIExtensions = append(se.UIExtensions, core.UIExtension{
		Name: "monms",
		FS: fstest.MapFS{
			"main.js": &fstest.MapFile{
				Data: adminMainJS,
				Mode: 0o644,
			},
		},
	})
}
