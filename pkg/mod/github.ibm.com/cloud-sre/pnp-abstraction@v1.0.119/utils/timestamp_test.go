package utils

import (
	"testing"
)

func TestVerifyAndConvertTimestamp(t *testing.T) {

	input1 := "2017-06-20 16:08:37-05"
	input2 := "2017-06-20T16:08:37-0500"
	input3 := "2017-06-20 23:08:37+02"
	input4 := "2017-06-21T01:38:37+0430"
	expected := "2017-06-20T21:08:37Z"
	input6 := "2019-08-27T15:29:37"	
	
	testInput(input1, expected, t)
	testInput(input2, expected, t)
	testInput(input3, expected, t)
	testInput(input4, expected, t)
	testInput(expected, expected, t)
	
	output,err := VerifyAndConvertTimestamp(input6)
	if err==nil {
		t.Error("<" + input6 + "> has unexpected format and shouldn't have been convertible. But was converted to <" + output + ">")
	}
}

func testInput(input string, expected string, t *testing.T) {
	output,err := VerifyAndConvertTimestamp(input)
	if err!=nil {
		t.Error("<" + input + "> should have been converted, but got error: ", err)
	} else if expected!=output {
		t.Error("<" + input + "> should have been converted to <" + expected + ">, but got <" + output + ">")
	}
}
