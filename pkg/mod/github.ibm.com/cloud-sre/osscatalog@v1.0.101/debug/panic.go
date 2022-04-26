package debug

import (
	"fmt"
	runtimeDebug "runtime/debug"
)

// recoveredErrorMsg is the error received from the last call to recover() (or empty string if there was no panic)
// We keep it as a global variable to avoid recursive panics
var recoveredErrorMsg string = ""

// PanicHandler prints information about a panic
// Note: PanicHandler must not call recover() itself, unless it is directly defined as a deferred function (not called *from* a deferred function)
func PanicHandler(recoveredObj interface{}) {
	if recoveredObj != nil {
		recoveredErrorMsg = fmt.Sprintf("PANIC: Recovered Error: %v\n", recoveredObj)
		recoveredErrorMsg += "Stack: " + string(runtimeDebug.Stack())
		Critical(recoveredErrorMsg)
	} else {
		recoveredErrorMsg = ""
	}
}

// GetPanicMessage return the error received from a panic (or empty string if there was no panic)
func GetPanicMessage() string {
	return recoveredErrorMsg
}
