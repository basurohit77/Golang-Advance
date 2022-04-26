// +build go1.13

package require

import "github.com/instana/testify/assert"

// ErrorAs asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
// This is a wrapper for errors.As.
func ErrorAs(t TestingT, err error, target interface{}, msgAndArgs ...interface{}) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if assert.ErrorAs(t, err, target, msgAndArgs...) {
		return
	}
	t.FailNow()
}

// ErrorAsf asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
// This is a wrapper for errors.As.
func ErrorAsf(t TestingT, err error, target interface{}, msg string, args ...interface{}) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if assert.ErrorAsf(t, err, target, msg, args...) {
		return
	}
	t.FailNow()
}

// ErrorIs asserts that at least one of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func ErrorIs(t TestingT, err error, target error, msgAndArgs ...interface{}) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if assert.ErrorIs(t, err, target, msgAndArgs...) {
		return
	}
	t.FailNow()
}

// ErrorIsf asserts that at least one of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func ErrorIsf(t TestingT, err error, target error, msg string, args ...interface{}) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if assert.ErrorIsf(t, err, target, msg, args...) {
		return
	}
	t.FailNow()
}

// NotErrorIs asserts that at none of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func NotErrorIs(t TestingT, err error, target error, msgAndArgs ...interface{}) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if assert.NotErrorIs(t, err, target, msgAndArgs...) {
		return
	}
	t.FailNow()
}

// NotErrorIsf asserts that at none of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func NotErrorIsf(t TestingT, err error, target error, msg string, args ...interface{}) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if assert.NotErrorIsf(t, err, target, msg, args...) {
		return
	}
	t.FailNow()
}
