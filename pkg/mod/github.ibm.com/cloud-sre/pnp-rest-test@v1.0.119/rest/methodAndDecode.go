package rest

import (
	"net/http"

	"github.ibm.com/cloud-sre/pnp-rest-test/common"
)

// GetAndDecode retrieves an api and decodes the returned object
func (s *Server) GetAndDecode(fct, objName, url string, intfce interface{}) error {
	METHOD := fct + "->[GetAndDecode]"

	resp, err := s.Get(METHOD, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return common.Decode(METHOD, objName, resp.Body, intfce)
}

// GetRequestAndDecode retrieves an api and decodes the returned object
func (s *Server) GetRequestAndDecode(fct, objName string, req *http.Request, intfce interface{}) error {
	METHOD := fct + "->[GetRequestAndDecode]"

	resp, err := s.GetRequest(METHOD, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return common.Decode(METHOD, objName, resp.Body, intfce)
}

// PostAndDecode retrieves an api and decodes the returned object
func (s *Server) PostAndDecode(fct, objName, url string, postBody string, intfce interface{}) error {
	METHOD := fct + "->[PostAndDecode]"

	resp, err := s.Post(METHOD, url, postBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return common.Decode(METHOD, objName, resp.Body, intfce)
}

// PostRequestAndDecode retrieves an api and decodes the returned object
func (s *Server) PostRequestAndDecode(fct, objName string, req *http.Request, intfce interface{}) error {
	METHOD := fct + "->[PostRequestAndDecode]"

	resp, err := s.PostRequest(METHOD, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return common.Decode(METHOD, objName, resp.Body, intfce)
}
