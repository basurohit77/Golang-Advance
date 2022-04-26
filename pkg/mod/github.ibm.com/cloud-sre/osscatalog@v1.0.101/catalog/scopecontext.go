package catalog

import (
	"context"
)

type scopeKeyType struct{}

var scopeKey scopeKeyType

// NewContextWithScope returns a new Context that contains the specified access scope
func NewContextWithScope(ctx context.Context, scope string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, scopeKey, scope)
}

// ScopeFromContext returns the access scope contained in this Context
func ScopeFromContext(ctx context.Context) (string, bool) {
	scope, ok := ctx.Value(scopeKey).(string)
	return scope, ok
}
