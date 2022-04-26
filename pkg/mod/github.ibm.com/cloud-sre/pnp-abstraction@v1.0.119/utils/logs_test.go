package utils

import (
	"testing"
	"time"
)

func TestSquelch(t *testing.T) {

	msg := "test message"
	squelch := LogSquelch(msg, time.Hour*1)

	if squelch {
		t.Fatal("Should not squelch")
	}

	squelch = LogSquelch(msg, time.Hour*1)

	if !squelch {
		t.Fatal("Should squelch")
	}
}
