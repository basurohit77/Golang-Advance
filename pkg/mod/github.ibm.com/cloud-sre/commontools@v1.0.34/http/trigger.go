package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type Metadata struct {
	URL     string
	Headers map[string]string
	Method  string
	Body    interface{}
	Cookies []*http.Cookie
	Timeout time.Duration // default: 10 minutes
}

type Result struct {
	Code   int //response code: e.g: 200, 401, 500...
	Data   []byte
	Cookie []*http.Cookie
	Err    error
}

//FireHTTPRequest http wrapper
func (m *Metadata) FireHTTPRequest() *Result {
	defaultTimeout := 10 * time.Minute
	if m.Timeout != 0 {
		defaultTimeout = m.Timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	r := &Result{
		Code:   400,
		Data:   nil,
		Cookie: nil,
		Err:    nil,
	}
	if m.URL == "" {
		r.Err = fmt.Errorf("url is required")
		return r
	}
	if m.Method == "" {
		r.Err = fmt.Errorf("method is required")
		return r
	}

	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}}

	var bodyReader io.Reader
	body := m.Body

	if body != nil {
		tp := reflect.TypeOf(body)
		tps := tp.String()
		if tps == "string" {
			bodyReader = strings.NewReader(body.(string))
		} else {
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				r.Err = err
				return r
			}
			bodyReader = strings.NewReader(string(bodyBytes))
		}
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(m.Method), m.URL, bodyReader)
	if err != nil {
		r.Err = err
		return r
	}

	for k, v := range m.Headers {
		if k == "BasicAuth" {
			usernamePwdPair := strings.Split(v, "::")
			if len(usernamePwdPair) == 2 {
				u := usernamePwdPair[0]
				p := usernamePwdPair[1]
				req.SetBasicAuth(u, p)
			}
			continue
		}
		req.Header.Set(k, v)
	}

	for _, cookie := range m.Cookies {
		req.AddCookie(cookie)
	}

	res, err := transport.RoundTrip(req)
	if err != nil {
		r.Err = err
		return r
	}

	if transport != nil {
		defer transport.CloseIdleConnections()
	}
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	var resCode int
	if res != nil && res.Body != nil {
		resCode = res.StatusCode
		data, err := ioutil.ReadAll(res.Body)
		r.Code = resCode
		r.Data = data
		r.Cookie = res.Cookies()
		r.Err = err
		return r
	}

	return r
}
