package catalog

import (
	"context"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// Token represents one IAM access token, suitable for creating an authentication Context for catalog operations
type Token string

type tokenKeyType struct{}

var tokenKey tokenKeyType
var tokenKeyRefreshable tokenKeyType // special support for storing a token that can be refreshed from the KeyFile

// NewContextWithToken returns a new Context that contains the specified IAM token
func NewContextWithToken(ctx context.Context, token Token) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := TokenFromContext(ctx); ok {
		panic("NewContextWithToken(): Context already contains a token")
	}
	return context.WithValue(ctx, tokenKey, token)
}

// NewContextWithTokenRefreshable returns a new Context that contains the specified IAM token,
// using a key from the KeyFile instead of the actual token, to allow it to be refreshed,
// This is only used in special cases where a series of calls (e.g. during ListEntries()) is so long
// that an externally-provided token might expire during the operation
// XXX This should never be called with a request coming from an external Client, as it would
// result in the Client using our internal token from the KeyFile instead of its own token
func NewContextWithTokenRefreshable(ctx context.Context, keyFileKey string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := TokenFromContext(ctx); ok {
		panic("NewContextWithTokenRefreshable(): Context already contains a token")
	}
	return context.WithValue(ctx, tokenKeyRefreshable, keyFileKey)
}

// TokenFromContext returns the IAM token contained in this Context
func TokenFromContext(ctx context.Context) (Token, bool) {

	// Common case: a non-refreshable token
	token, ok := ctx.Value(tokenKey).(Token)
	if ok {
		return token, true
	}

	// Check if there is a special KeyFile based token
	keyFileKey, ok := ctx.Value(tokenKeyRefreshable).(string)
	if ok {
		token, err := rest.GetToken(keyFileKey)
		if err != nil {
			debug.PrintError("Cannot get IAM token for Global Catalog from KeyFile in Context: %v", err)
			return "", false
		}
		return Token(token), true
	}

	return "", false
}
