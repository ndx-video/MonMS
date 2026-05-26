package testutil

import (
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// NewSuperuser creates a superuser record for integration tests.
func NewSuperuser(t *testing.T, app core.App, email string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("find superusers collection: %v", err)
	}

	rec := core.NewRecord(collection)
	rec.Set("email", email)
	rec.SetPassword("password123456")
	if err := app.Save(rec); err != nil {
		t.Fatalf("save superuser: %v", err)
	}
	return rec
}

// AuthClient returns an HTTP client that sends Bearer auth for the given superuser.
func AuthClient(t *testing.T, app core.App, user *core.Record) *http.Client {
	t.Helper()

	token, err := user.NewAuthToken()
	if err != nil {
		t.Fatalf("new auth token: %v", err)
	}

	return &http.Client{
		Transport: bearerTransport{token: token},
	}
}
