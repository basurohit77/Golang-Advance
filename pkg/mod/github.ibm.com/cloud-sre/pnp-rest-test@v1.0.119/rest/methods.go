package rest

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Get performs REST GET request
func (s *Server) Get(fct, url string) (*http.Response, error) {
	METHOD := fct + "->[rest.Get]"

	log.Println(METHOD, "Requesting:", url)
	tTraceStart := time.Now()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if s.Token != "" {
		req.Header["Authorization"] = []string{s.Token}
	} else if s.SNToken != "" {
		req.Header["Authorization"] = []string{"Bearer " + s.SNToken}
	} else if s.User != "" && s.Password != "" {
		req.SetBasicAuth(s.User, s.Password)
	}

	client := &http.Client{}
	if os.Getenv("disableSkipTransport") != "true" {
		tr := &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyFromEnvironment,
		}

		if strings.Index(url, "https") == 0 {
			client.Transport = tr
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(METHOD, "Request failed:", err.Error())
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errMsg := "Received non-2XX return code: " + resp.Status
		log.Println(METHOD, errMsg)
		return nil, errors.New(errMsg)
	}

	tTraceEnd := time.Now()
	log.Println(METHOD, "Request time:", tTraceEnd.Sub(tTraceStart).String())

	return resp, err
}

// GetRequest performs REST GET request
func (s *Server) GetRequest(fct string, req *http.Request) (*http.Response, error) {
	METHOD := fct + "->[rest.Get]"

	log.Println(METHOD, "Requesting:", req.URL.String())
	tTraceStart := time.Now()

	client := &http.Client{}

	if os.Getenv("disableSkipTransport") != "true" {
		tr := &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		if strings.Index(req.URL.String(), "https") == 0 {
			client.Transport = tr
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(METHOD, "Request failed:", err.Error())
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errMsg := "Received non-2XX return code: " + resp.Status
		log.Println(METHOD, errMsg)
		return nil, errors.New(errMsg)
	}

	tTraceEnd := time.Now()
	log.Println(METHOD, "Request time:", tTraceEnd.Sub(tTraceStart).String())

	return resp, err
}

// Post performs REST POST request
func (s *Server) Post(fct, url string, postBody string) (*http.Response, error) {
	METHOD := fct + "->[rest.Post]"

	log.Println(METHOD, "Requesting:", url)
	tTraceStart := time.Now()

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(postBody)))
	if err != nil {
		return nil, err
	}

	if s.Token != "" {
		req.Header["Authorization"] = []string{s.Token}
	} else if s.SNToken != "" {
		req.Header["Authorization"] = []string{"Bearer " + s.SNToken}
	} else if s.User != "" && s.Password != "" {
		req.SetBasicAuth(s.User, s.Password)
	}

	client := &http.Client{}

	if os.Getenv("disableSkipTransport") != "true" {
		tr := &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		if strings.Index(url, "https") == 0 {
			client.Transport = tr
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(METHOD, "Request failed:", err.Error())
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(METHOD, "Failed to read response body:", err.Error())
		} else {
			log.Println(METHOD, "Response body:", string(data))
		}

		errMsg := "Received non-2XX return code: " + resp.Status
		log.Println(METHOD, errMsg)
		return nil, errors.New(errMsg)
	}

	tTraceEnd := time.Now()
	log.Println(METHOD, "Request time:", tTraceEnd.Sub(tTraceStart).String())

	return resp, err
}

// PostRequest performs REST POST request
func (s *Server) PostRequest(fct string, req *http.Request) (*http.Response, error) {
	METHOD := fct + "->[rest.Post]"

	log.Println(METHOD, "Requesting:", req.URL.String())
	tTraceStart := time.Now()

	client := &http.Client{}

	if os.Getenv("disableSkipTransport") != "true" {
		tr := &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		if strings.Index(req.URL.String(), "https") == 0 {
			client.Transport = tr
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(METHOD, "Request failed:", err.Error())
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errMsg := "Received non-2XX return code: " + resp.Status
		log.Println(METHOD, errMsg)
		return nil, errors.New(errMsg)
	}

	tTraceEnd := time.Now()
	log.Println(METHOD, "Request time:", tTraceEnd.Sub(tTraceStart).String())

	return resp, err
}

// Delete performs REST DELETE request
func (s *Server) Delete(fct, url string) (*http.Response, error) {
	METHOD := fct + "->[rest.Delete]"

	log.Println(METHOD, "Requesting:", url)
	tTraceStart := time.Now()

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	if s.Token != "" {
		req.Header["Authorization"] = []string{s.Token}
	} else if s.SNToken != "" {
		req.Header["Authorization"] = []string{"Bearer " + s.SNToken}
	} else if s.User != "" && s.Password != "" {
		req.SetBasicAuth(s.User, s.Password)
	}

	client := &http.Client{}

	if os.Getenv("disableSkipTransport") != "true" {

		tr := &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		if strings.Index(url, "https") == 0 {
			client.Transport = tr
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(METHOD, "Request failed:", err.Error())
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errMsg := "Received non-2XX return code: " + resp.Status
		log.Println(METHOD, errMsg)
		return nil, errors.New(errMsg)
	}

	tTraceEnd := time.Now()
	log.Println(METHOD, "Request time:", tTraceEnd.Sub(tTraceStart).String())

	return resp, err
}

// Patch performs REST PATCH request
func (s *Server) Patch(fct, url string, patchBody string) (*http.Response, error) {
	METHOD := fct + "->[rest.Patch]"

	log.Println(METHOD, "Requesting:", url)
	tTraceStart := time.Now()

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer([]byte(patchBody)))
	if err != nil {
		return nil, err
	}

	if s.Token != "" {
		req.Header["Authorization"] = []string{s.Token}
	} else if s.SNToken != "" {
		req.Header["Authorization"] = []string{"Bearer " + s.SNToken}
	} else if s.User != "" && s.Password != "" {
		req.SetBasicAuth(s.User, s.Password)
	}

	client := &http.Client{}

	if os.Getenv("disableSkipTransport") != "true" {

		tr := &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		if strings.Index(url, "https") == 0 {
			client.Transport = tr
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(METHOD, "Request failed:", err.Error())
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(METHOD, "Failed to read response body:", err.Error())
		} else {
			log.Println(METHOD, "Response body:", string(data))
		}

		errMsg := "Received non-2XX return code: " + resp.Status
		log.Println(METHOD, errMsg)
		return nil, errors.New(errMsg)
	}

	tTraceEnd := time.Now()
	log.Println(METHOD, "Request time:", tTraceEnd.Sub(tTraceStart).String())

	return resp, err
}
