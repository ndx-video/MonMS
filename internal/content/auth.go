package content

import (
	"crypto/subtle"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// RequirePublishToken gates content import routes with MONMS_PUBLISH_TOKEN (PUB-05).
// Fails closed when expected is empty (production token unset).
func RequirePublishToken(expected string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if expected == "" {
			return e.UnauthorizedError("invalid publish token", nil)
		}

		auth := e.Request.Header.Get("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			return e.UnauthorizedError("invalid publish token", nil)
		}
		return e.Next()
	}
}
