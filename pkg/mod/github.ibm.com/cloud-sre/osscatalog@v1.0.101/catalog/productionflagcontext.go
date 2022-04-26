package catalog

import (
	"context"
)

// Special flag to control access to Production data
type productionFlag string

const productionFlagReadWrite productionFlag = "enabled, read-write, Picard, authorization Alpha-Alpha-3-0-5"
const productionFlagReadOnly productionFlag = "enabled, read-only"
const productionFlagDisabled productionFlag = "disabled"

type productionKeyType struct{}

var productionKey productionKeyType

// NewContextStaging returns a new Context for (read-write) access to the Staging environment
func NewContextStaging(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, productionKey, productionFlagDisabled)
}

// NewContextProductionReadOnly returns a new Context for read-only access to the Production environment
func NewContextProductionReadOnly(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, productionKey, productionFlagReadOnly)
}

// NewContextProductionReadWrite returns a new Context for read-write access to the Production environment
func NewContextProductionReadWrite(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, productionKey, productionFlagReadWrite)
}

// ProductionFlagFromContext determines if this context allows access to the Production environment (or Staging), and read-write (or read-only)
func ProductionFlagFromContext(ctx context.Context) (production bool, readwrite bool, ok bool) {
	flag, ok := ctx.Value(productionKey).(productionFlag)
	if !ok {
		return false, false, false /* dummy values */
	}
	switch flag {
	case productionFlagReadWrite:
		return true, true, true
	case productionFlagReadOnly:
		return true, false, true
	case productionFlagDisabled:
		return false, true, true /* always read-write for Staging */
	default:
		return false, false, false /* dummy values */
	}
}
