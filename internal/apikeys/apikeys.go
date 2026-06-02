package apikeys

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monms/monms/internal/content"
	"github.com/pocketbase/pocketbase/core"
)

// ResolvedKey ties an API key record to its owning auth record.
type ResolvedKey struct {
	KeyRecord *core.Record
	Owner     *core.Record
}

// Generate creates a new API key secret and its stored metadata (hash + prefix).
func Generate() (secret string, prefix string, hash string, err error) {
	raw := make([]byte, secretLen)
	if _, err = rand.Read(raw); err != nil {
		return "", "", "", err
	}
	secret = tokenPrefix + hex.EncodeToString(raw)
	if len(secret) < prefixLen {
		return "", "", "", fmt.Errorf("api key shorter than prefix length")
	}
	prefix = secret[:prefixLen]
	hash = HashSecret(secret, "")
	return secret, prefix, hash, nil
}

// HashSecret returns SHA-256 hex of pepper+secret.
func HashSecret(secret, pepper string) string {
	sum := sha256.Sum256([]byte(pepper + secret))
	return hex.EncodeToString(sum[:])
}

// Pepper returns the API key hashing pepper from env or a site-stable fallback.
func Pepper(siteAbs string) string {
	if p := strings.TrimSpace(os.Getenv("MONMS_API_KEY_PEPPER")); p != "" {
		return p
	}
	if siteAbs == "" {
		return "monms-default-pepper"
	}
	sum := sha256.Sum256([]byte(filepath.Clean(siteAbs)))
	return hex.EncodeToString(sum[:16])
}

// IsSuperuserAuth reports whether the auth record is a PocketBase superuser.
func IsSuperuserAuth(auth *core.Record) bool {
	return auth != nil && auth.Collection().Name == core.CollectionNameSuperusers
}

// CanManageKeys reports whether the session may create/list/revoke API keys in the dashboard.
func CanManageKeys(auth *core.Record, cfg content.MonmsConfig) bool {
	if auth == nil {
		return false
	}
	if IsSuperuserAuth(auth) {
		return true
	}
	return cfg.MCP.AllowNonSuperuserKeys && auth.Collection().Name == CollectionUsers
}

// OwnerMatchesAuth reports whether keyRecord belongs to auth.
func OwnerMatchesAuth(keyRecord, auth *core.Record) bool {
	if keyRecord == nil || auth == nil {
		return false
	}
	if IsSuperuserAuth(auth) {
		return keyRecord.GetString("superuser") == auth.Id
	}
	if auth.Collection().Name == CollectionUsers {
		return keyRecord.GetString("user") == auth.Id
	}
	return false
}

// SetOwnerFields sets superuser or user relation on a new key record.
func SetOwnerFields(rec *core.Record, owner *core.Record) error {
	if owner == nil {
		return fmt.Errorf("owner required")
	}
	switch owner.Collection().Name {
	case core.CollectionNameSuperusers:
		rec.Set("superuser", owner.Id)
		rec.Set("user", "")
	case CollectionUsers:
		rec.Set("user", owner.Id)
		rec.Set("superuser", "")
	default:
		return fmt.Errorf("unsupported owner collection %q", owner.Collection().Name)
	}
	return nil
}

// OwnerFromKey returns the auth record referenced by the API key.
func OwnerFromKey(app core.App, keyRecord *core.Record) (*core.Record, error) {
	if id := keyRecord.GetString("superuser"); id != "" {
		return app.FindRecordById(core.CollectionNameSuperusers, id)
	}
	if id := keyRecord.GetString("user"); id != "" {
		return app.FindRecordById(CollectionUsers, id)
	}
	return nil, fmt.Errorf("api key has no owner")
}

// Create stores a new API key for owner and returns the one-time secret.
func Create(app core.App, siteAbs string, owner *core.Record, name string) (secret string, rec *core.Record, err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, fmt.Errorf("name required")
	}

	secret, prefix, _, err := Generate()
	if err != nil {
		return "", nil, err
	}

	coll, err := app.FindCollectionByNameOrId(CollectionName)
	if err != nil {
		return "", nil, err
	}

	rec = core.NewRecord(coll)
	rec.Set("name", name)
	rec.Set("prefix", prefix)
	rec.Set("secretHash", HashSecret(secret, Pepper(siteAbs)))
	if err := SetOwnerFields(rec, owner); err != nil {
		return "", nil, err
	}
	if err := app.Save(rec); err != nil {
		return "", nil, err
	}
	return secret, rec, nil
}

// Resolve validates bearerToken and returns the key and owner records.
func Resolve(app core.App, siteAbs, bearerToken string) (*ResolvedKey, error) {
	bearerToken = strings.TrimSpace(bearerToken)
	if !strings.HasPrefix(bearerToken, tokenPrefix) {
		return nil, fmt.Errorf("invalid api key format")
	}

	pepper := Pepper(siteAbs)
	hash := HashSecret(bearerToken, pepper)
	prefix := bearerToken
	if len(prefix) > prefixLen {
		prefix = prefix[:prefixLen]
	}

	coll, err := app.FindCollectionByNameOrId(CollectionName)
	if err != nil {
		return nil, err
	}

	records, err := app.FindRecordsByFilter(
		CollectionName,
		"prefix = {:prefix}",
		"-id",
		50,
		0,
		map[string]any{"prefix": prefix},
	)
	if err != nil {
		return nil, err
	}

	var matched *core.Record
	for _, rec := range records {
		stored := rec.GetString("secretHash")
		if subtle.ConstantTimeCompare([]byte(stored), []byte(hash)) == 1 {
			matched = rec
			break
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("invalid api key")
	}

	owner, err := OwnerFromKey(app, matched)
	if err != nil {
		return nil, err
	}

	_ = coll // keep coll referenced for future rule checks

	return &ResolvedKey{KeyRecord: matched, Owner: owner}, nil
}

// TouchLastUsed updates lastUsedAt on the key record.
func TouchLastUsed(app core.App, keyRecord *core.Record) {
	keyRecord.Set("lastUsedAt", time.Now().UTC().Format("2006-01-02 15:04:05.000Z"))
	_ = app.Save(keyRecord)
}

// ListForOwner returns API key records owned by auth (no secret material).
func ListForOwner(app core.App, auth *core.Record) ([]*core.Record, error) {
	var filter string
	var params map[string]any
	if IsSuperuserAuth(auth) {
		filter = "superuser = {:id}"
		params = map[string]any{"id": auth.Id}
	} else if auth.Collection().Name == CollectionUsers {
		filter = "user = {:id}"
		params = map[string]any{"id": auth.Id}
	} else {
		return nil, fmt.Errorf("unsupported auth collection")
	}
	return app.FindRecordsByFilter(CollectionName, filter, "-id", 200, 0, params)
}

// Revoke deletes an API key if it belongs to auth.
func Revoke(app core.App, auth *core.Record, keyID string) error {
	rec, err := app.FindRecordById(CollectionName, keyID)
	if err != nil {
		return err
	}
	if !OwnerMatchesAuth(rec, auth) {
		return fmt.Errorf("forbidden")
	}
	return app.Delete(rec)
}

// OwnerEmail returns the owner email when available.
func OwnerEmail(owner *core.Record) string {
	if owner == nil {
		return ""
	}
	return owner.GetString("email")
}

// IsPublisherOwner reports publisher allowlist membership for the key owner.
func IsPublisherOwner(owner *core.Record, cfg content.MonmsConfig) bool {
	return content.IsPublisher(OwnerEmail(owner), cfg.PublisherEmails)
}
