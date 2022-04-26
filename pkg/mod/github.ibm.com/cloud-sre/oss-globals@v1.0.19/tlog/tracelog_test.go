package tlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {

	var str string = Log()
	assert.Contains(t, str, "TestLog")
}

func TestFuncName(t *testing.T) {
	var str string = FuncName()
	assert.Equal(t, str, "TestFuncName")
}
