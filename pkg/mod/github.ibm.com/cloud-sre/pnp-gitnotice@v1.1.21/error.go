package gitnotice

import "fmt"

// Error is an advanced error structure that can carry external messages
type Error struct {
	Err         []error
	ExternalMsg []string
}

// NewError will create a new error
func NewError(err error, externalMsg string, inserts ...interface{}) *Error {

	result := &Error{Err: []error{err}, ExternalMsg: []string{externalMsg}}

	if len(inserts) > 0 {
		result.ExternalMsg[0] = fmt.Sprintf(externalMsg, inserts...)
	}

	return result
}

// Meld will take two errors and combine them into one error
func (e *Error) Meld(second *Error) *Error {
	first := e

	if first == nil {
		return second
	}
	if second == nil {
		return first
	}

	first.Err = append(first.Err, second.Err...)
	first.ExternalMsg = append(first.ExternalMsg, second.ExternalMsg...)
	return first
}

// Add will add additional error information to the given error
func (e *Error) Add(err error, externalMsg string, inserts ...interface{}) *Error {
	second := NewError(err, externalMsg, inserts...)
	return e.Meld(second)
}

// String will format an external message for an error
func (e *Error) String() string {

	msg := ""

	if len(e.Err) == 0 || len(e.ExternalMsg) != len(e.Err) {
		return ""
	}

	msgCount := len(e.ExternalMsg)
	tab := ""

	if msgCount > 1 {
		msg += fmt.Sprintf("%d Messages: \n", msgCount)
		tab = "   "
	}

	for i, m := range e.ExternalMsg {

		if msgCount > 1 {
			msg += fmt.Sprintf("%s%d. ", tab, i+1)
		}

		if len(m) > 0 {
			msg += m
			if len(e.Err) > i && e.Err[i] != nil {
				msg += fmt.Sprintf(" (Details: %s)\n", e.Err[i].Error())
			}
		} else if len(e.Err) > i && e.Err[i] != nil {
			msg += fmt.Sprintf("%s\n", e.Err[i].Error())
		}

	}

	return msg
}
