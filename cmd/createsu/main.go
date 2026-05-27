package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	siteDir := "/tmp/monms-test/site-stage"
	os.Setenv("MONMS_SITE", siteDir)

	app := pocketbase.New()
	app.RootCmd.SetArgs([]string{"serve", "--dir", filepath.Join(siteDir, ".pb_data")})

	if err := app.Bootstrap(); err != nil {
		fmt.Println("ERR bootstrap:", err)
		os.Exit(1)
	}

	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		fmt.Println("ERR find superusers:", err)
		os.Exit(1)
	}

	rec := core.NewRecord(superusers)
	rec.Set("email", "admin@monms.test")
	rec.SetPassword("admin123")
	if err := app.Save(rec); err != nil {
		fmt.Println("ERR save:", err)
		os.Exit(1)
	}

	fmt.Println("OK superuser created:", rec.Id)
}