// +build go1.13

package assert

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

func ExampleErrorAssertionFunc() {
	t := &testing.T{} // provided by test

	dumbParseNum := func(input string, v interface{}) error {
		return json.Unmarshal([]byte(input), v)
	}

	tests := []struct {
		name      string
		arg       string
		assertion ErrorAssertionFunc
	}{
		{"1.2 is number", "1.2", NoError},
		{"1.2.3 not number", "1.2.3", Error},
		{"true is not number", "true", Error},
		{"3 is number", "3", NoError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var x float64
			tt.assertion(t, dumbParseNum(tt.arg, &x))
		})
	}
}

func TestErrorAs(t *testing.T) {
	mockT := new(testing.T)
	tests := []struct {
		err    error
		result bool
	}{
		{fmt.Errorf("wrap: %w", &customError{}), true},
		{io.EOF, false},
		{nil, false},
	}
	for _, tt := range tests {
		tt := tt
		var target *customError
		t.Run(fmt.Sprintf("ErrorAs(%#v,%#v)", tt.err, target), func(t *testing.T) {
			res := ErrorAs(mockT, tt.err, &target)
			if res != tt.result {
				t.Errorf("ErrorAs(%#v,%#v) should return %t)", tt.err, target, tt.result)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	mockT := new(testing.T)
	tests := []struct {
		err    error
		target error
		result bool
	}{
		{io.EOF, io.EOF, true},
		{fmt.Errorf("wrap: %w", io.EOF), io.EOF, true},
		{io.EOF, io.ErrClosedPipe, false},
		{nil, io.EOF, false},
		{io.EOF, nil, false},
		{nil, nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("ErrorIs(%#v,%#v)", tt.err, tt.target), func(t *testing.T) {
			res := ErrorIs(mockT, tt.err, tt.target)
			if res != tt.result {
				t.Errorf("ErrorIs(%#v,%#v) should return %t", tt.err, tt.target, tt.result)
			}
		})
	}
}

func TestNotErrorIs(t *testing.T) {
	mockT := new(testing.T)
	tests := []struct {
		err    error
		target error
		result bool
	}{
		{io.EOF, io.EOF, false},
		{fmt.Errorf("wrap: %w", io.EOF), io.EOF, false},
		{io.EOF, io.ErrClosedPipe, true},
		{nil, io.EOF, true},
		{io.EOF, nil, true},
		{nil, nil, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("NotErrorIs(%#v,%#v)", tt.err, tt.target), func(t *testing.T) {
			res := NotErrorIs(mockT, tt.err, tt.target)
			if res != tt.result {
				t.Errorf("NotErrorIs(%#v,%#v) should return %t", tt.err, tt.target, tt.result)
			}
		})
	}
}
