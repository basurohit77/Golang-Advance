package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFireHTTPRequest(t *testing.T) {
	assert := assert.New(t)
	input := "test"
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(strings.ToUpper(input)))
		}),
	)
	defer s.Close()

	m := &Metadata{
		URL:     s.URL,
		Headers: nil,
		Method:  "GET",
		Body:    nil,
		Cookies: nil,
		Timeout: 0,
	}
	result := m.FireHTTPRequest()
	assert.Equal(200, result.Code, "code should be 200")
	assert.NoError(result.Err, "should no error")
	assert.Equal(strings.ToUpper(input), string(result.Data))
	assert.Empty(result.Cookie, "no cookie")
}

func TestFireHTTPRequestTimeout(t *testing.T) {
	assert := assert.New(t)
	input := "test"
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(5 * time.Second)
			w.Write([]byte(strings.ToUpper(input)))
		}),
	)
	defer s.Close()

	m := &Metadata{
		URL:     s.URL,
		Headers: nil,
		Method:  "GET",
		Body:    nil,
		Cookies: nil,
		Timeout: 2 * time.Second,
	}
	result := m.FireHTTPRequest()
	assert.Equal(400, result.Code, "code should be 400")
	assert.EqualError(result.Err, "context deadline exceeded")
	assert.Equal([]uint8([]byte(nil)), result.Data)
	assert.Empty(result.Cookie, "no cookie")
}

func TestFireHTTPRequestWrongURL(t *testing.T) {
	assert := assert.New(t)

	m := &Metadata{
		URL:     "http://127.0.0.1",
		Headers: nil,
		Method:  "GET",
		Body:    nil,
		Cookies: nil,
	}
	result := m.FireHTTPRequest()
	assert.Equal(400, result.Code, "code should be 400")
	assert.Greater(result.Err.Error(), "connection refused")
	assert.Equal([]uint8([]byte(nil)), result.Data)
	assert.Empty(result.Cookie, "no cookie")
}

func TestFireHTTPRequestTLS(t *testing.T) {
	assert := assert.New(t)

	m := &Metadata{
		URL:     "https://iam.cloud.ibm.com/identity/.well-known/openid-configuration",
		Headers: nil,
		Method:  "GET",
		Body:    nil,
		Cookies: nil,
	}
	result := m.FireHTTPRequest()
	assert.Equal(200, result.Code, "code should be 200")
	assert.NoError(result.Err)
	assert.NotEmpty(result.Data)
	assert.Empty(result.Cookie, "no cookie")
}

func TestFireHTTPRequestNotTrustTLS(t *testing.T) {
	assert := assert.New(t)
	input := "test"
	s := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(5 * time.Second)
			w.Write([]byte(strings.ToUpper(input)))
		}),
	)
	defer s.Close()
	m := &Metadata{
		URL:     s.URL,
		Headers: nil,
		Method:  "GET",
		Body:    nil,
		Cookies: nil,
	}
	result := m.FireHTTPRequest()
	assert.Equal(400, result.Code, "code should be 400")
	assert.EqualError(result.Err, "x509: certificate signed by unknown authority", "must check the tls")
	assert.Empty(result.Data)
	assert.Empty(result.Cookie, "cookie returned")
}

func TestFireHTTPRequestWWWFormURLEncoded(t *testing.T) {
	assert := assert.New(t)
	input := "test"
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := r.Header
			for k, v := range headers {
				log.Println(k, v)
			}
			b, err := ioutil.ReadAll(r.Body)
			assert.Empty(err)
			log.Print(string(b))
			w.Write([]byte(input))
		}),
	)
	defer s.Close()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	data := url.Values{}
	data.Set("grant_type", "urn:ibm:params:oauth:grant-type:kkk")
	data.Set("response_type", "bar")
	data.Set("kkk", "123456")
	m := &Metadata{
		URL:     s.URL,
		Headers: headers,
		Method:  "POST",
		Body:    data.Encode(),
	}
	result := m.FireHTTPRequest()
	assert.Equal(200, result.Code, "code should be 200")
	assert.Empty(result.Err, "")
	assert.Equal(input, string(result.Data))
	assert.Empty(result.Cookie, "no cookie")
}
