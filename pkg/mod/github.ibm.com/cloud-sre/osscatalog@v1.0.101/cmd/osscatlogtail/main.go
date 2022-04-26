package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// AppVersion is the current version number for the osscatlogtail program
const AppVersion = "0.1"

func main() {
	var versionFlag = flag.Bool("version", false, "Print the version number of this program")
	var keyFile = flag.String("keyfile", "", "Path to file containing authentication keys")
	var levelFlag = flag.String("level", "AUDIT", "Levels of log to return (or ALL)")
	var fromFlag = flag.String("from", "1d", "How many days/hours from the current time go back into the logs")
	var prettyFlag = flag.Bool("pretty", false, "Pretty-print the output")

	flag.Parse()

	os.Stderr.WriteString(fmt.Sprintf("%s Version: %s\n" /*os.Args[0]*/, "osscatlogtail", AppVersion))
	if *versionFlag {
		return
	}

	if *keyFile == "" || strings.ToLower(*keyFile) == "default" {
		err := rest.LoadDefaultKeyFile()
		if err != nil {
			panic(fmt.Sprintf("Error reading DEFAULT keyfile: %v", err))
		}
	} else {
		err := rest.LoadKeyFile(*keyFile)
		if err != nil {
			panic(fmt.Sprintf("Error reading keyfile: %v", err))
		}
	}

	var levels string
	if strings.ToLower(*levelFlag) == "all" {
		levels = ""
	} else {
		levels = *levelFlag
	}
	var from int64
	to := time.Now().UnixNano() / int64(time.Millisecond)
	if *fromFlag != "" {
		var days int64
		_, err := fmt.Sscanf(*fromFlag, "%dd", &days)
		if err == nil {
			from = to - (days * 24 * int64(time.Hour/time.Millisecond))
		} else {
			duration, err := time.ParseDuration(*fromFlag)
			if err == nil {
				from = to - int64(duration/time.Millisecond)
			} else {
				panic(fmt.Sprintf("Invalid duration in \"-from\" parameter: %v", err))
			}
		}
	} else {
		from = to - (24 * int64(time.Hour/time.Millisecond))
	}

	timeLocation, _ := time.LoadLocation("UTC")
	fromString := time.Unix(from/1000, 0).In(timeLocation).Format("2006-01-02T1504Z")
	os.Stderr.WriteString(fmt.Sprintf("Getting osscat logs since %s (levels=%s)\n", fromString, *levelFlag))

	err := ExportLogDNA(from, to, levels, *prettyFlag, os.Stdout)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

// ExportLogDNA exports LogDNA content to a Writer
func ExportLogDNA(from, to int64, levels string, pretty bool, w io.Writer) error {
	timeLocation, _ := time.LoadLocation("UTC")
	client := &http.Client{}
	requestURL, err := url.Parse(debug.LogDNAExportURL)
	if err != nil {
		return debug.WrapError(err, "Cannot parse base URL for exporting LogDNA content")
	}
	key, err := rest.GetKey(debug.LogDNAServiceKeyName)
	if err != nil {
		return debug.WrapError(err, "Cannot find the service key for LogDNA")
	}
	requestURL.User = url.UserPassword(key, "")
	requestValues := url.Values{}
	requestValues.Set("from", strconv.FormatInt(from, 10))
	requestValues.Set("to", strconv.FormatInt(to, 10))
	//	requestValues.Set("size", "10000")
	if levels != "" {
		requestValues.Set("levels", levels)
	}
	requestURL.RawQuery = requestValues.Encode()
	resp, err := client.Get(requestURL.String())
	if err != nil {
		return debug.WrapError(err, "Error exporting LogDNA content")
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				return fmt.Errorf("HTTP error exporting LogDNA content: %v / %v - cannot read response body: %v", resp.StatusCode, resp.Status, err2)
			}
			return fmt.Errorf("HTTP error exporting LogDNA content: %v / %v - response body: %v", resp.StatusCode, resp.Status, string(body))
		}
	} else {
		return fmt.Errorf("Null response while exporting LogDNA content")
	}

	if pretty {
		dec := json.NewDecoder(resp.Body)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		for dec.More() {
			data := make(map[string]interface{})
			err = dec.Decode(&data)
			if err != nil {
				return debug.WrapError(err, "Error decoding JSON output from LogDNA for pretty-printing")
			}
			var TIMESTAMP string
			if ts, found := data["_ts"]; found {
				if ts1, ok := ts.(float64); ok {
					TIMESTAMP = time.Unix(int64(ts1)/1000, 0).In(timeLocation).Format("2006-01-02T1504Z")
				} else {
					TIMESTAMP = fmt.Sprintf(`***Error: cannot convert "_ts=%v" to float64`, ts)
				}
			} else {
				TIMESTAMP = fmt.Sprintf(`***Error: "_ts" attribute not found`)
			}
			fmt.Fprintf(w, `"%s": `, TIMESTAMP)
			err = enc.Encode(data)
			if err != nil {
				return debug.WrapError(err, "Error re-encoding JSON output from LogDNA for pretty-printing")
			}
		}
	} else {
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			return debug.WrapError(err, "Error copying raw JSON output from LogDNA")
		}
	}

	return nil
}
