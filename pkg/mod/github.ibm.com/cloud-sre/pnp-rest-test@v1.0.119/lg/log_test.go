package lg

import (
	"errors"
	"testing"
)

func TestLg1(t *testing.T) {
	msg := getLogMsg("ERROR", "FUNCTION", errors.New("ERRMSG"), "Hello %s", "world")
	if msg != "[ERROR][FUNCTION] Hello world (error:ERRMSG)" {
		t.Fatal("Did not get expected message. Got:" + msg)
	}
}
