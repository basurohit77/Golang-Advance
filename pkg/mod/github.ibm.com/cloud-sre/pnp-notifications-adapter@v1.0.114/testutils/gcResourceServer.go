package testutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
)

const (
	level0Server = "globalcatalog/testdata/resources_test_%d.json"
	level1Server = "../globalcatalog/testdata/resources_test_%d.json"
)

var gTS *httptest.Server

// GetGCResourceServer will return a server that can provide Global Catalog Resources
func GetGCResourceServer(t *testing.T) (server *httptest.Server, page1 string) {
	return gcResourceServer(t, level1Server)
}

// GetGCResourceServerL0 will return a server that can provide Global Catalog Resources using UT data from level 0
func GetGCResourceServerL0(t *testing.T) (server *httptest.Server, page1 string) {
	return gcResourceServer(t, level0Server)
}

func gcResourceServer(t *testing.T, fn string) (server *httptest.Server, page1 string) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		// We know the last character of the URL will be the index for the data file to use
		index, err := strconv.Atoi(r.URL.Path[len(r.URL.Path)-1:])
		if err != nil {
			t.Fatal("Error creating index:", err)
		}

		fn := fmt.Sprintf(fn, index)
		// Remove the ReadFile below because this is a unit test library only. Needs to be in a normaly file because Go cannot share UT libraries
		data, err := ioutil.ReadFile(filepath.Clean(fn)) // File in source directory with test file
		if err != nil {
			t.Fatal("Error reading file:", err)
		}

		data = bytes.Replace(data, []byte("NEXT_LINK_TEST_PLACEHOLDER"), []byte(fmt.Sprintf("%s/%d", gTS.URL, index+1)), -1)
		_, err = w.Write(data)

		if err != nil {
			t.Fatal("Error reading file:", err)
		}
	}))

	// Need this for the handler function later
	gTS = ts

	return ts, fmt.Sprintf("%s/1", ts.URL)
}
