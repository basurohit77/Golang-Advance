package shared

import (
	"os"
	"testing"
)

func TestHasValidToken(t *testing.T) {

	if ok := HasValidToken(os.Getenv("SNOW_TOKEN")); !ok {
		t.Fatal(ok)
	}

}

func TestHasInvalidToken(t *testing.T) {

	if ok := HasValidToken("some invalid token"); ok {
		t.Fatal(ok)
	}

}
