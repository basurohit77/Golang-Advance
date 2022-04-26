package utils

import (
	"os"
	"testing"
	"time"
)

func TestGetEnvMin(t *testing.T) {
	d := GetEnvMinutes("TestEnvVariable", 15)
	if d != time.Minute*15 {
		t.Fatal("Did not get default value, got", d)
	}

	os.Setenv("TestEnvVariable", "180")
	d = GetEnvMinutes("TestEnvVariable", 15)
	if d != time.Minute*180 {
		t.Fatal("Did not get specific value. Got", d)
	}
}
