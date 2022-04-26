// +build go1.13

package require

// ErrorAs asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
// This is a wrapper for errors.As.
func (a *Assertions) ErrorAs(err error, target interface{}, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	ErrorAs(a.t, err, target, msgAndArgs...)
}

// ErrorAsf asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
// This is a wrapper for errors.As.
func (a *Assertions) ErrorAsf(err error, target interface{}, msg string, args ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	ErrorAsf(a.t, err, target, msg, args...)
}

// ErrorIs asserts that at least one of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func (a *Assertions) ErrorIs(err error, target error, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	ErrorIs(a.t, err, target, msgAndArgs...)
}

// ErrorIsf asserts that at least one of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func (a *Assertions) ErrorIsf(err error, target error, msg string, args ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	ErrorIsf(a.t, err, target, msg, args...)
}

// NotErrorIs asserts that at none of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func (a *Assertions) NotErrorIs(err error, target error, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	NotErrorIs(a.t, err, target, msgAndArgs...)
}

// NotErrorIsf asserts that at none of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func (a *Assertions) NotErrorIsf(err error, target error, msg string, args ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	NotErrorIsf(a.t, err, target, msg, args...)
}
