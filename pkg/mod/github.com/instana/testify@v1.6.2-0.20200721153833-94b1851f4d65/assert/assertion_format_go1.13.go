// +build go1.13

package assert

// ErrorIsf asserts that at least one of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func ErrorIsf(t TestingT, err error, target error, msg string, args ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	return ErrorIs(t, err, target, append([]interface{}{msg}, args...)...)
}

// ErrorAsf asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
// This is a wrapper for errors.As.
func ErrorAsf(t TestingT, err error, target interface{}, msg string, args ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	return ErrorAs(t, err, target, append([]interface{}{msg}, args...)...)
}

// NotErrorIsf asserts that at none of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func NotErrorIsf(t TestingT, err error, target error, msg string, args ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	return NotErrorIs(t, err, target, append([]interface{}{msg}, args...)...)
}
