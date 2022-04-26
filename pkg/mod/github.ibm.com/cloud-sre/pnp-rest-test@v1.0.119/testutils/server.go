package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// DataServer is a function type to return data
type DataServer func(url *url.URL) (data []byte, httpStatus int)

// NewDataServer allows you to pass a function pointer for a function to return data
func NewDataServer(srvr DataServer) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, code := srvr(r.URL)

		if code != http.StatusOK {
			w.WriteHeader(code)
		}
		w.Header().Set("Content-Type", "application/json")

		if len(data) > 0 {
			w.Write(data)
		}
	}))

	return ts
}

// NewServer creates a new test server instance
func NewServer(data []byte) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))

	return ts
}

// NewRdrServer creates a new test server instance
func NewRdrServer(rdr io.Reader) *httptest.Server {

	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Fatal("Could not read for test server", err)
	}
	return NewServer(data)
}

// NewErrorServer creates a new test server instance that returns an error
func NewErrorServer(statusCode int) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	}))

	return ts
}

// NewJSONServer creates a new test server to serve JSON
func NewJSONServer(obj interface{}) *httptest.Server {

	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(obj); err != nil {
		log.Fatal("ERROR could not encode object")
	}

	return NewServer(buffer.Bytes())
}
