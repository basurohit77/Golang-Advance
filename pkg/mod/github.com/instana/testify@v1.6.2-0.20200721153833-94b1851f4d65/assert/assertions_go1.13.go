// +build go1.13

package assert

import (
	"errors"
	"fmt"
)

// ErrorIs asserts that at least one of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func ErrorIs(t TestingT, err, target error, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if errors.Is(err, target) {
		return true
	}

	var expectedText string
	if target != nil {
		expectedText = target.Error()
	}

	chain := buildErrorChainString(err)

	return Fail(t, fmt.Sprintf("Target error should be in err chain:\n"+
		"expected: %q\n"+
		"in chain: %s", expectedText, chain,
	), msgAndArgs...)
}

// NotErrorIs asserts that at none of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func NotErrorIs(t TestingT, err, target error, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if !errors.Is(err, target) {
		return true
	}

	var expectedText string
	if target != nil {
		expectedText = target.Error()
	}

	chain := buildErrorChainString(err)

	return Fail(t, fmt.Sprintf("Target error should not be in err chain:\n"+
		"found: %q\n"+
		"in chain: %s", expectedText, chain,
	), msgAndArgs...)
}

// ErrorAs asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
// This is a wrapper for errors.As.
func ErrorAs(t TestingT, err error, target interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if errors.As(err, target) {
		return true
	}

	chain := buildErrorChainString(err)

	return Fail(t, fmt.Sprintf("Should be in error chain:\n"+
		"expected: %q\n"+
		"in chain: %s", target, chain,
	), msgAndArgs...)
}

func buildErrorChainString(err error) string {
	if err == nil {
		return ""
	}

	e := errors.Unwrap(err)
	chain := fmt.Sprintf("%q", err.Error())
	for e != nil {
		chain += fmt.Sprintf("\n\t%q", e.Error())
		e = errors.Unwrap(e)
	}
	return chain
}
