package gitnotice

import (
	"errors"
	"strings"
	"testing"
)

func TestError(t *testing.T) {

	err1Msg := "something broke"

	ext1Msg := "Hello World"

	err2Msg := "another msg"
	ext2Msg := "an insert %s ok"

	err := NewError(errors.New(err1Msg), ext1Msg)
	err.Meld(NewError(errors.New(err2Msg), ext2Msg, "foobar"))

	checkMsg := `2 Messages: 
   1. Hello World (Details: something broke)
   2. an insert foobar ok (Details: another msg)
`

	if checkMsg != err.String() {
		t.Log("Messages don't match")
		t.Log("WAS :" + strings.ReplaceAll(err.String(), "\n", "\\n"))
		t.Fatal("WANT:" + strings.ReplaceAll(checkMsg, "\n", "\\n"))
	}
}
