package testutils

import (
	"net/http"
	"testing"
)

func TestGc(t *testing.T) {
	ts, p1 := GetGCResourceServer(t)
	defer ts.Close()

	http.Get(p1)

}
