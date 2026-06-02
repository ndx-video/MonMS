package authbootstrap

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterBootstrapHook ensures engine auth collections exist and wires API key validation.
func RegisterBootstrapHook(app core.App) {
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if err := ensureUsersCollection(e.App); err != nil {
			return err
		}
		return ensureAPIKeysCollection(e.App)
	})

	registerAPIKeyHooks(app)
}

func ensureUsersCollection(app core.App) error {
	if _, err := app.FindCollectionByNameOrId(CollectionUsers); err == nil {
		return nil
	}

	c := core.NewAuthCollection(CollectionUsers)
	// No public self-registration; superusers manage users via PocketBase admin.
	c.ListRule = nil
	c.ViewRule = nil
	c.CreateRule = nil
	c.UpdateRule = nil
	c.DeleteRule = nil

	return app.Save(c)
}

func ensureAPIKeysCollection(app core.App) error {
	if _, err := app.FindCollectionByNameOrId(CollectionAPIKeys); err == nil {
		return nil
	}

	usersCol, err := app.FindCollectionByNameOrId(CollectionUsers)
	if err != nil {
		return fmt.Errorf("authbootstrap: users collection: %w", err)
	}
	superCol, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return fmt.Errorf("authbootstrap: superusers collection: %w", err)
	}

	c := core.NewBaseCollection(CollectionAPIKeys)
	c.ListRule = nil
	c.ViewRule = nil
	c.CreateRule = nil
	c.UpdateRule = nil
	c.DeleteRule = nil

	c.Fields.Add(
		&core.TextField{Name: "name", Required: true, Min: 1, Max: 120},
		&core.TextField{Name: "prefix", Required: true, Min: 8, Max: 16},
		&core.TextField{Name: "secretHash", Required: true, Min: 64, Max: 64},
		&core.RelationField{
			Name:         "superuser",
			MaxSelect:    1,
			CollectionId: superCol.Id,
		},
		&core.RelationField{
			Name:         "user",
			MaxSelect:    1,
			CollectionId: usersCol.Id,
		},
		&core.DateField{Name: "lastUsedAt"},
	)

	return app.Save(c)
}

func registerAPIKeyHooks(app core.App) {
	app.OnRecordCreate(CollectionAPIKeys).BindFunc(func(e *core.RecordEvent) error {
		if err := validateAPIKeyOwner(e.Record); err != nil {
			return err
		}
		return e.Next()
	})
	app.OnRecordUpdate(CollectionAPIKeys).BindFunc(func(e *core.RecordEvent) error {
		if err := validateAPIKeyOwner(e.Record); err != nil {
			return err
		}
		return e.Next()
	})
}

func validateAPIKeyOwner(rec *core.Record) error {
	hasSuper := rec.GetString("superuser") != ""
	hasUser := rec.GetString("user") != ""
	if hasSuper == hasUser {
		return fmt.Errorf("exactly one of superuser or user must be set")
	}
	return nil
}
