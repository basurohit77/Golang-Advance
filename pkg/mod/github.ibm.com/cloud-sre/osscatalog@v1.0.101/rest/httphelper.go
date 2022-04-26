package rest

// httphelper.go contains utility functions for making HTTP REST calls

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/md5" // #nosec G501
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/debug"
)

const httpTimeout = 120

var disableTLSVerifyFlag bool

// SetDisableTLSVerify sets a flag to temporarily disable TLS security verifications in operations in this package
// It returns the previous value of the flag -- intended for in "defer" statements
// XXX Note that this is not thread-safe
func SetDisableTLSVerify(newValue bool) bool {
	old := disableTLSVerifyFlag
	disableTLSVerifyFlag = newValue
	return old
}

// getBaseContentType extracts the base content-type information (e.g. "application/json") from the http headers
func getBaseContentType(header http.Header) string {
	ctype := header.Get("Content-Type")
	ctypeIx := strings.Index(ctype, ";") // strip the charset and/our boundary parts
	if ctypeIx != -1 {
		ctype = ctype[:ctypeIx]
	}
	return ctype
}

// DoHTTPGet performs all the common steps for doing a GET from some REST API and unmarshalling the results
func DoHTTPGet(fullURL string, authorization string, headers http.Header, label string, debugFlag int, result interface{}) error {
	var client *http.Client
	if disableTLSVerifyFlag { // #nosec G402
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr, Timeout: time.Second * httpTimeout}
	} else {
		client = &http.Client{Timeout: time.Second * httpTimeout}
	}

	request, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		err = debug.WrapError(err, "Error creating HTTP request for %s  (url=%s)", label, fullURL)
		return err
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Add("Accept-Encoding", "gzip, deflate")
	//	request.Header.Add("Accept-Encoding", "gzip")
	//	request.Header.Add("Accept-Encoding", "deflate")
	for k, vals := range headers {
		if len(vals) > 1 {
			for _, v := range vals {
				request.Header.Add(k, v)
			}
		} else {
			request.Header.Set(k, vals[0])
		}
	}
	debug.Debug(debugFlag, "Fetching one entry from %s: http.Get(%s)", label, fullURL)
	start := time.Now()
	resp, err := client.Do(request)
	if resp != nil {
		defer resp.Body.Close()
		debug.Debug(debugFlag, "%s --> StatusCode=%v  %+v   %+v\nresult  encoding: Content-Encoding=%v   TransferEncoding=%v   Uncompressed=%v",
			label, resp.StatusCode, resp.Status, resp,
			resp.Header.Get("Content-Encoding"), resp.TransferEncoding, resp.Uncompressed)
	}
	err = MakeHTTPError(err, resp, false, "Error in HTTP GET for %s  (url=%s)", label, fullURL)
	recordTiming(label, err, start, time.Now())
	if err != nil {
		return err
	}
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			err = debug.WrapError(err, "Error creating a gzip reader to uncompress response from %s  (url=%s)", label, fullURL)
			return err
		}
		defer reader.Close()
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		err = debug.WrapError(err, "Error reading response body for %s  (url=%s)", label, fullURL)
		return err
	}
	switch getBaseContentType(resp.Header) {
	case "application/json":
		if debug.IsDebugEnabled(debugFlag | debug.Fine) /* || true /* XXX */ {
			//		var jsonmap compare.MapOfInterfaces
			var jsonmap interface{}
			err = json.Unmarshal(body, &jsonmap)
			if err != nil {
				err = debug.WrapError(err, "Error in json.Unmarshal to map for %s  (url=%s)", label, fullURL)
				debug.Debug(debugFlag|debug.Fine, "HTTP response body: %s", string(body))
				return err
			}
			str2, _ := json.MarshalIndent(jsonmap, "DEBUG: ", "   ")
			debug.Debug(debugFlag|debug.Fine, "%s", "*** Raw JSON object returned from GET\n"+string(str2))
			//str3, _ := json.MarshalIndent(result, "DEBUG ", "   ")
			//debug.Debug(debugFlag|debug.Fine, "result=\nDEBUG %s", str3)
		}

		err = json.Unmarshal(body, result)
		if err != nil {
			err = debug.WrapError(err, "Error in json.Unmarshal for %s  (url=%s)", label, fullURL)
			if debug.IsDebugEnabled(debugFlag) {
				debug.Debug(debugFlag, "HTTP response body: %s", string(body))
			}
			return err
		}

		if debug.IsDebugEnabled(debugFlag | debug.Fine) {
			//		var jsonmap compare.MapOfInterfaces
			var jsonmap interface{}
			err = json.Unmarshal(body, &jsonmap)
			if err != nil {
				err = debug.WrapError(err, "Error in json.Unmarshal to map for %s  (url=%s)", label, fullURL)
				return err
			}
			out := compare.Output{IncludeEqual: true}
			compare.DeepCompare("result", result, "json", &jsonmap, &out)
			str := out.StringWithPrefix("DEBUG:")
			debug.Debug(debugFlag|debug.Fine, "%s", "*** Comparing parsed result with raw JSON object returned from GET\n"+str)
		}

	case "application/xml", "text/xml":
		if debug.IsDebugEnabled(debugFlag | debug.Fine) /* || true /* XXX */ {
			debug.Debug(debugFlag, "HTTP response body: %s", string(body))
		}
		err = xml.Unmarshal(body, result)
		if err != nil {
			err = debug.WrapError(err, "Error in xml.Unmarshal for %s  (url=%s)", label, fullURL)
			if debug.IsDebugEnabled(debugFlag) {
				debug.Debug(debugFlag, "HTTP response body: %s", string(body))
			}
			return err
		}

	default:
		err = debug.WrapError(err, "Unexpected content type %s in response for %s  (url=%s)", resp.Header.Get("Content-Type"), label, fullURL)
		return err
	}

	return nil
}

// DoHTTPPostOrPut performs all the common steps for doing a POST or PUT to some REST API
func DoHTTPPostOrPut(operation string, fullURL string, authorization string, headers http.Header, data interface{}, result interface{}, label string, debugFlag int) error {
	var client *http.Client
	if disableTLSVerifyFlag { // #nosec G402
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr, Timeout: time.Second * httpTimeout}
	} else {
		client = &http.Client{Timeout: time.Second * httpTimeout}
	}

	var req io.Reader
	var contentType string
	var request *http.Request
	var err error
	var contentHash [md5.Size]byte
	if data != nil {
		switch v := data.(type) {
		case url.Values:
			encoded := v.Encode()
			req = bytes.NewBufferString(encoded)
			contentType = "application/x-www-form-urlencoded"
		//	case struct{}, *struct{}:
		default:
			switch getBaseContentType(headers) {
			case "application/xml":
				buf, err := xml.Marshal(v)
				if err != nil {
					err = debug.WrapError(err, "Error marshaling struct input data into XML for HTTP request for %s  (url=%s)", label, fullURL)
					return err
				}
				req = bytes.NewBuffer(buf)
				contentType = "application/xml"
				contentHash = md5.Sum(buf) // #nosec G401
			default:
				buf, err := json.Marshal(v)
				if err != nil {
					err = debug.WrapError(err, "Error marshaling struct input data into JSON for HTTP request for %s  (url=%s)", label, fullURL)
					return err
				}
				req = bytes.NewBuffer(buf)
				contentType = "application/json"
			}
			//	default:
			//		panic(fmt.Sprintf("DoHTTPPostOrPut() does not implement input data type %T", data))
		}
		request, err = http.NewRequest(operation, fullURL, req)
	} else {
		request, err = http.NewRequest(operation, fullURL, nil)
	}
	if err != nil {
		err = debug.WrapError(err, "Error creating HTTP request for %s  (url=%s)", label, fullURL)
		return err
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	request.Header.Set("Accept", "application/json")
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	for k, vals := range headers {
		if len(vals) > 1 {
			for _, v := range vals {
				request.Header.Add(k, v)
			}
		} else {
			request.Header.Set(k, vals[0])
		}
	}
	if contentHash[0] != 0 {
		request.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(contentHash[:]))
	}
	debug.Debug(debugFlag, "Writing one entry to %s: http.%s(%s)", label, operation, fullURL)
	debug.Debug(debugFlag|debug.Fine, "request=%v", request)

	start := time.Now()
	resp, err := client.Do(request)
	if resp != nil {
		defer resp.Body.Close()
		debug.Debug(debugFlag, "%s --> StatusCode=%v  %v   %#v", label, resp.StatusCode, resp.Status, resp)
	}
	err = MakeHTTPError(err, resp, false, "Error in HTTP %s for %s  (url=%s)", operation, label, fullURL)
	recordTiming(label, err, start, time.Now())
	if err != nil {
		return err
	}
	debug.Debug(debugFlag, "result encoding: Content-Encoding=%v   TransferEncoding=%v   Uncompressed=%v", resp.Header.Get("Content-Encoding"), resp.TransferEncoding, resp.Uncompressed)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = debug.WrapError(err, "Error reading response body for %s  (url=%s)", label, fullURL)
		return err
	}

	if len(body) > 0 {
		if result == nil {
			result = new(interface{})
		}
		// Note: this may overwrite the result if there is an error response
		switch getBaseContentType(resp.Header) {
		case "application/json":
			if debug.IsDebugEnabled(debugFlag | debug.Fine) /* || true /* XXX */ {
				debug.Debug(debugFlag, "HTTP response body: %s", string(body))
			}
			err = json.Unmarshal(body, result)
			if err != nil {
				err = debug.WrapError(err, "Error in json.Unmarshal for %s  (url=%s)", label, fullURL)
				if debug.IsDebugEnabled(debugFlag) {
					debug.Debug(debugFlag, "HTTP response body: %s", string(body))
				}
				return err
			}

			if debug.IsDebugEnabled(debugFlag) {
				str, _ := json.MarshalIndent(result, "DEBUG ", "   ")
				debug.Debug(debugFlag, "result=\nDEBUG %s", str)
			}
		case "application/xml", "text/xml":
			if debug.IsDebugEnabled(debugFlag | debug.Fine) /* || true /* XXX */ {
				debug.Debug(debugFlag, "HTTP response body: %s", string(body))
			}
			err = xml.Unmarshal(body, result)
			if err != nil {
				err = debug.WrapError(err, "Error in xml.Unmarshal for %s  (url=%s)", label, fullURL)
				if debug.IsDebugEnabled(debugFlag) {
					debug.Debug(debugFlag, "HTTP response body: %s", string(body))
				}
				return err
			}

		default:
			err = debug.WrapError(err, "Unexpected content type %s in response for %s  (url=%s)", resp.Header.Get("Content-Type"), label, fullURL)
			return err
		}
	}
	return nil
}

type timingRecord struct {
	duration     time.Duration
	count        int
	failDuration time.Duration
	failCount    int
}

var timings = make(map[string]*timingRecord)
var timingMutex = &sync.Mutex{}

func recordTiming(label string, err error, start, end time.Time) {
	timingMutex.Lock()

	tr := timings[label]
	if tr == nil {
		tr = &timingRecord{}
		timings[label] = tr
	}
	d := end.Sub(start)
	tr.duration += d
	tr.count++
	if err != nil {
		tr.failDuration += d
		tr.failCount++
	}
	// We do not use a deferred call here because there is only one clear path through this short routine; might as well avoid some overhead
	timingMutex.Unlock()
}

// PrintTimings generates a log message that summarize all the time spent in all REST calls,
// organized by label
func PrintTimings() {
	timingMutex.Lock()
	defer timingMutex.Unlock()
	buffer := strings.Builder{}
	labels := make([]string, 0, len(timings))
	for l := range timings {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	for _, l := range labels {
		tr := timings[l]
		buffer.WriteString(fmt.Sprintf("        %-22s: %4d calls / %8.3fs       (errors: %4d calls / %8.3fs)\n", l, tr.count, tr.duration.Seconds(), tr.failCount, tr.failDuration.Seconds()))
	}
	debug.Info("Time spent in all REST calls:\n%s", buffer.String())
}
