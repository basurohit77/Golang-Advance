package catalog

import (
	"context"
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// setupContextForMainEntries sets up a Context suitable for accessing for access Main entries in Global catalog
// The Context includes a set of flags indicating we should access the Production or Staging Catalog,
// whether to access in read-only or read-write mode, a scope, and a IAM token.
// The "refreshable" option allows to store a key from a KeyFile instead of a plain IAM token, which makes it possible to be refreshed.
func setupContextForMainEntries(prod productionFlag, refreshable bool) (ctx context.Context, err error) {
	var tokenKey string
	switch prod {
	case productionFlagReadWrite:
		ctx = NewContextProductionReadWrite(nil)
		tokenKey = catalogMainKeyName
	case productionFlagReadOnly:
		ctx = NewContextProductionReadOnly(nil)
		tokenKey = catalogMainKeyName
	case productionFlagDisabled:
		ctx = NewContextStaging(nil)
		tokenKey = catalogOSSKeyNameStaging // XXX no separate key for Main entries in Staging
	default:
		panic(fmt.Sprintf("Invalid productionFlag: %s", prod))
	}

	var scope string
	switch options.GlobalOptions().VisibilityRestrictions {
	case catalogapi.VisibilityPrivate:
		scope = "&account=global"
	case catalogapi.VisibilityIBMOnly:
		scope = ""
	case catalogapi.VisibilityPublic:
		scope = ""
		tokenKey = "" // Use unauthenticated access
	default:
		panic(fmt.Sprintf("Unknown value for GlobalOptions().VisibilityRestrictions: %v", options.GlobalOptions().VisibilityRestrictions))
	}
	ctx = NewContextWithScope(ctx, scope)

	if tokenKey != "" {
		if refreshable {
			ctx = NewContextWithTokenRefreshable(ctx, tokenKey)
		} else {
			token, err := rest.GetToken(tokenKey)
			if err != nil {
				err = debug.WrapError(err, "Cannot get IAM token for Global Catalog (Main entries)")
				return nil, err
			}
			ctx = NewContextWithToken(ctx, Token(token))
		}
	} else {
		// Note that we allow an empty token -- to support unauthenticated requests (e.g to see public Catalog entries only)
		ctx = NewContextWithToken(ctx, "")
	}

	return ctx, nil
}

// setupContextForOSSEntries sets up a Context suitable for accessing OSS entries in Global catalog
// The Context includes a set of flags indicating we should access the Production or Staging Catalog,
// whether to access in read-only or read-write mode, a scope, and a IAM token.
// The scope is usually ignored when accessing OSS entries, but is present here for consistency with setupContextForMainEntries()
func setupContextForOSSEntries(prod productionFlag) (ctx context.Context, err error) {
	var tokenKey string

	switch prod {
	case productionFlagReadWrite:
		ctx = NewContextProductionReadWrite(nil)
		tokenKey = catalogOSSKeyNameProduction
	case productionFlagReadOnly:
		ctx = NewContextProductionReadOnly(nil)
		tokenKey = catalogOSSKeyNameProduction
	case productionFlagDisabled:
		ctx = NewContextStaging(nil)
		tokenKey = catalogOSSKeyNameStaging
	default:
		panic(fmt.Sprintf("Invalid productionFlag: %s", prod))
	}

	// We do not set a Scope context for OSS entries

	token, err := rest.GetToken(tokenKey)
	if err != nil {
		err = debug.WrapError(err, "Cannot get IAM token for Global Catalog (OSS entries)")
		return nil, err
	}
	ctx = NewContextWithToken(ctx, Token(token))

	return ctx, nil
}

func readContextForMainEntries(ctx context.Context, readwrite bool) (url string, scope string, token Token, err error) {
	token, ok := TokenFromContext(ctx)
	if !ok {
		err := fmt.Errorf("Context object for Catalog access does not specify a IAM token (even an empty one) (Main entries)")
		// We do not want to panic here, but we cannot allow the user to proceed if they did not explicitly request read-write access
		return "", "", "", err
	}
	scope, ok = ScopeFromContext(ctx)
	if !ok {
		debug.Debug(debug.Catalog, "No scope specified in the Context object, assuming empty scope (Main entries)")
		scope = ""
	}
	prod, rwctx, ok := ProductionFlagFromContext(ctx)
	if !ok {
		debug.Debug(debug.Catalog, "No production flag specified in the Context object, assuming ProductionReadOnly (Main entries)")
		prod = true
		rwctx = false
	}
	if prod {
		if readwrite && !rwctx {
			err := fmt.Errorf("Context object for Production Catalog access does not specify read-write access (Main entries)")
			// We do not want to panic here, but we cannot allow the user to proceed if they did not explicitly request read-write access
			return "", "", "", err
		}
		if readwrite && !options.IsProductionWriteEnabled() {
			err := fmt.Errorf("Attempting write access to Production Global Catalog but special option is not enabled (Main entries)")
			return "", "", "", err
		}
		return catalogMainURL, scope, token, nil
	}
	return catalogOSSURLStaging, scope, token, nil // XXX no separate URL for Main entries in Staging
}

func readContextForOSSEntries(ctx context.Context, readwrite bool) (url string, scope string, token Token, err error) {
	token, ok := TokenFromContext(ctx)
	if !ok {
		err := fmt.Errorf("Context object for Catalog access does not specify a IAM token (even an empty one) (OSS entries)")
		// We do not want to panic here, but we cannot allow the user to proceed if they did not explicitly request read-write access
		return "", "", "", err
	}
	scope, ok = ScopeFromContext(ctx)
	if !ok {
		debug.Debug(debug.Catalog, "No scope specified in the Context object, assuming empty scope (irrelevant for OSS entries)")
		scope = ""
	}
	prod, rwctx, ok := ProductionFlagFromContext(ctx)
	if !ok {
		debug.Debug(debug.Catalog, "No production flag specified in the Context object, assuming ProductionReadOnly (OSS entries)")
		prod = true
		rwctx = false
	}
	if prod {
		if readwrite && !rwctx {
			err := fmt.Errorf("Context object for Production Catalog access does not specify read-write access (OSS entries)")
			// We do not want to panic here, but we cannot allow the user to proceed if they did not explicitly request read-write access
			return "", "", "", err
		}
		if readwrite && !options.IsProductionWriteEnabled() {
			err := fmt.Errorf("Attempting write access to Production Global Catalog but special option is not enabled (OSS entries)")
			return "", "", "", err
		}
		return catalogOSSURLProduction, scope, token, nil
	}
	return catalogOSSURLStaging, scope, token, nil
}

func getCatalogNameFromContext(ctx context.Context) string {
	prod, _, ok := ProductionFlagFromContext(ctx)
	if !ok {
		debug.Debug(debug.Catalog, "No production flag specified in the Context object, assuming ProductionReadOnly (OSS entries)")
		prod = true
	}
	if prod {
		return "Production"
	}
	return "Staging"
}
