package debug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Functions for forwarding log information to LogDNA

// LogLevel represents the log level (INFO, ERROR, etc.) for a LogEntry
type LogLevel string

// Constant values for LogLevel
const (
	LevelINFO     LogLevel = "INFO"
	LevelERROR    LogLevel = "ERROR"
	LevelWARN     LogLevel = "WARN"
	LevelAUDIT    LogLevel = "AUDIT"
	LevelCRITICAL LogLevel = "CRITICAL"
)

// LogEntry is the structure of one log entry to be sent to LogDNA
type LogEntry struct {
	Timestamp int64    `json:"timestamp"`
	Level     LogLevel `json:"level"` // INFO, ERROR, AUDIT, etc.
	App       string   `json:"app"`
	Line      string   `json:"line"`
	Meta      struct {
		OSSEntryID string `json:"ossentryid"` // Optional: the OSS entry to which this log entry relates
	} `json:"meta"`
}

type logDNAClientType struct {
	mutex          sync.Mutex
	ingestionKey   string
	hostname       string
	app            string
	panicOnError   bool
	maxBufferLines int
	maxBufferSize  int
	bufferSize     int
	buffer         []*LogEntry
	httpClient     *http.Client
}

var logDNAClient *logDNAClientType

// SetupLogDNA initializes the logging to LogDNA for this instance of the library
//
// Note: we cannot obtain the ingestion key directly from the keyfile inside this function,
// because this would introduce a circular dependency between packages
func SetupLogDNA(ingestionKey string, app string, panicOnError bool, maxLines int) (err error) {
	if logDNAClient != nil {
		panic("Duplicate call to SetupLogDNA()")
	}
	defer func() {
		if err != nil {
			if logDNAClient != nil && logDNAClient.panicOnError {
				logDNAClient = nil
				panic(err)
			} else {
				logDNAClient = nil
			}
		}
	}()
	logDNAClient = &logDNAClientType{
		ingestionKey:  ingestionKey,
		app:           app,
		maxBufferSize: 12000,
		panicOnError:  panicOnError,
	}
	if maxLines == 0 {
		logDNAClient.maxBufferLines = 100
	} else {
		logDNAClient.maxBufferLines = maxLines
	}
	logDNAClient.buffer = make([]*LogEntry, 0, logDNAClient.maxBufferLines)
	logDNAClient.bufferSize = 0
	logDNAClient.hostname, err = os.Hostname()
	if err != nil {
		return WrapError(err, "Cannot get hostname while setting-up logging to LogDNA")
	}
	transport := &http.Transport{
		//		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 30 * time.Second,
	}
	logDNAClient.httpClient = &http.Client{Transport: transport}

	err = sendToLogDNA(LevelINFO, false, "", "Initialized oss-catalog logging")
	if err != nil {
		return WrapError(err, "Error while testing initial connection to LogDNA (writing test log entry)")
	}
	err = FlushLogDNA()
	if err != nil {
		return WrapError(err, "Error while testing initial connection to LogDNA (flushing test log entry)")
	}

	return nil
}

func sendToLogDNA(level LogLevel, flush bool, id string, msg string) error {
	if logDNAClient != nil {
		e := LogEntry{
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			Level:     level,
			App:       logDNAClient.app,
			Line:      msg + ".",
		}
		e.Meta.OSSEntryID = id
		newSize := len(e.Line)
		if newSize > logDNAClient.maxBufferSize {
			e.Line = e.Line[:(logDNAClient.maxBufferSize-20)] + "...((truncated))"
			newSize = len(e.Line)
		}
		if flush {
			// Flush before and after, to ensure that the message stands out in the output and comes out immediately
			err := FlushLogDNA()
			if err != nil {
				return err
			}
		}
		logDNAClient.mutex.Lock()
		var numTries int
		// Need a loop here because we release the mutex before FlushLogDNA() and someone else might fill the buffer again before we come back
		for (len(logDNAClient.buffer) >= logDNAClient.maxBufferLines) || ((logDNAClient.bufferSize + newSize) >= logDNAClient.maxBufferSize) {
			logDNAClient.mutex.Unlock()
			if numTries > 10 {
				return fmt.Errorf("Possible infinite loop trying to flush LogDNA buffer (%d tries)", numTries)
			}
			numTries++
			err := FlushLogDNA()
			if err != nil {
				return nil
			}
			logDNAClient.mutex.Lock()
		}
		logDNAClient.buffer = append(logDNAClient.buffer, &e)
		logDNAClient.bufferSize += newSize
		if IsDebugEnabled(LogDNA) {
			if len(msg) > 20 {
				Debug(LogDNA, "sendToLogDNA() buffer lines=%d  buffer size=%d  msg=\"%s\" ...", len(logDNAClient.buffer), logDNAClient.bufferSize, msg[:19])
			} else {
				Debug(LogDNA, "sendToLogDNA() buffer lines=%d  buffer size=%d  msg=\"%s\"", len(logDNAClient.buffer), logDNAClient.bufferSize, msg)
			}
		}
		logDNAClient.mutex.Unlock()
		if flush {
			// Flush before and after, to ensure that the message stands out in the output and comes out immediately
			err := FlushLogDNA()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var maskURLRegex = regexp.MustCompile(`https://.*@`)

// FlushLogDNA flushes all writes to LogDNA
func FlushLogDNA() (err error) {
	if logDNAClient != nil {
		logDNAClient.mutex.Lock()
		if len(logDNAClient.buffer) == 0 {
			logDNAClient.mutex.Unlock()
			return nil
		}
		defer func() {
			if err != nil {
				if logDNAClient != nil && logDNAClient.panicOnError {
					logDNAClient = nil
					panic(err)
				}
			}
		}()
		Debug(LogDNA, "FlushLogDNA(): sending %d lines  (approx %d bytes) to LogDNA", len(logDNAClient.buffer), logDNAClient.bufferSize)
		body := struct {
			Lines []*LogEntry `json:"lines"`
		}{
			Lines: logDNAClient.buffer,
		}
		logDNAClient.buffer = make([]*LogEntry, 0, logDNAClient.maxBufferLines)
		logDNAClient.bufferSize = 0
		logDNAClient.mutex.Unlock()
		var jsonBytes []byte
		jsonBytes, err = json.Marshal(body)
		if err != nil {
			return WrapError(err, "Cannot marshal JSON body for flushing LogDNA output")
		}
		Debug(LogDNA|Fine, "FlushLogDNA(): payload: %s", string(jsonBytes))
		jsonReader := bytes.NewReader(jsonBytes)
		var resp *http.Response
		var requestURL *url.URL
		requestURL, err = url.Parse(LogDNAIngestionURL)
		if err != nil {
			return WrapError(err, "Cannot parse base URL for flushing LogDNA output")
		}
		requestURL.User = url.UserPassword(logDNAClient.ingestionKey, "")
		requestValues := url.Values{}
		requestValues.Set("hostname", logDNAClient.hostname)
		requestValues.Set("now", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))
		requestURL.RawQuery = requestValues.Encode()
		for retryCount := 0; retryCount < 3; retryCount++ {
			resp, err = logDNAClient.httpClient.Post(requestURL.String(), "application/json; charset=UTF-8", jsonReader)
			if err == nil {
				break
			}
			cleanErr := maskURLRegex.ReplaceAllString(err.Error(), `https://****@`) // To mask the ingestion key embedded in the URL
			Warning("Retry flushing LogDNA output after error: %v", cleanErr)
		}
		if err != nil {
			return WrapError(err, "Error flushing LogDNA output")
		}
		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, err2 := ioutil.ReadAll(resp.Body)
				if err2 != nil {
					return fmt.Errorf("HTTP error flushing LogDNA output: %v / %v - cannot read response body: %v", resp.StatusCode, resp.Status, err2)
				}
				return fmt.Errorf("HTTP error flushing LogDNA output: %v / %v - response body: %v", resp.StatusCode, resp.Status, string(body))
			}
			if IsDebugEnabled(LogDNA) {
				body, _ := ioutil.ReadAll(resp.Body)
				Debug(LogDNA, "FlushLogDNA() server returned %v / %v - response body: %v", resp.StatusCode, resp.Status, string(body))
			}
		} else {
			return fmt.Errorf("Null response while flushing LogDNA output")
		}
	}
	return nil
}
