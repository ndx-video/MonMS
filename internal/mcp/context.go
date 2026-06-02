package mcp

import (
	"context"

	"github.com/monms/monms/internal/apikeys"
)

type ctxKey struct{}

type session struct {
	Resolved   *apikeys.ResolvedKey
	OwnerToken string
}

func withSession(ctx context.Context, s *session) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

func sessionFromContext(ctx context.Context) (*session, bool) {
	s, ok := ctx.Value(ctxKey{}).(*session)
	return s, ok
}
