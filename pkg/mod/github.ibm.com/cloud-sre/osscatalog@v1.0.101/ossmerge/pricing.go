package ossmerge

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// PricingInfo records the part numbers for one Catalog Entry
type PricingInfo struct {
	CatalogID   string                           `json:"catalog_id"`
	CatalogName string                           `json:"catalog_name"`
	CatalogKind string                           `json:"kind"`
	CatalogPath string                           `json:"catalog_path"`
	PartNumbers []string                         `json:"part_numbers"`
	Issues      []*ossvalidation.ValidationIssue `json:"issues"`
}

// JSON returns a JSON-style string representation of this PricingInfo record
func (pi *PricingInfo) JSON() string {
	var result strings.Builder
	result.WriteString("   ")
	json, _ := json.MarshalIndent(pi, "    ", "    ")
	_, err := result.Write(json)
	if err != nil {
		panic(err)
	}
	return result.String()
}

var allPricingInfo map[string]*PricingInfo

// LoadPricingFile loads pricing info from a side file
func LoadPricingFile(filename string) (numEntries int, err error) {
	file, err := os.Open(filename) // #nosec G304
	if err != nil {
		return 0, debug.WrapError(err, "Cannot open Pricing input file %s", filename)
	}
	defer file.Close() // #nosec G307

	allPricingInfo = make(map[string]*PricingInfo)

	dec := json.NewDecoder(file)
	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return 0, debug.WrapError(err, "Error reading initial token from Pricing input file")
		}
		if token0, ok := token.(json.Delim); !ok || token0.String() != `[` {
			return 0, fmt.Errorf(`Expected initial token "[" in Pricing input file, got "%#v"`, token)
		}
		for dec.More() {
			var e *PricingInfo
			err := dec.Decode(&e)
			if err != nil {
				details := parseJSONError(err)
				return 0, debug.WrapError(err, "Error reading entry from Pricing input file: %s", details)
			}
			if e.CatalogID == "" {
				// Ignore empty entry
				continue
			}
			if previous, found := allPricingInfo[e.CatalogID]; !found {
				allPricingInfo[e.CatalogID] = e
				debug.Debug(debug.Fine, "Recording Pricing entry %+v", e)
			} else if e == previous {
				debug.Warning("Ignoring duplicate entry from Pricing input file: %+v", e)
			} else {
				debug.PrintError("Ignoring conflicting entry from Pricing input file: %+v  (previous=%+v)", e, previous)
			}
		}
		token, err = dec.Token()
		if err != nil {
			return numEntries, debug.WrapError(err, "Error reading final token from Pricing input file")
		}
		if token0, ok := token.(json.Delim); !ok || token0.String() != `]` {
			return numEntries, fmt.Errorf(`Expected final token "]" in Pricing input file, got "%#v"`, token)
		}
	}

	/*
		rawData, err := ioutil.ReadAll(file)
		if err != nil {
			return 0, debug.WrapError(err, "Error reading Pricing input file %s", filename)
		}
		var input interface{}
		err = json.Unmarshal(rawData, &input)
		if err != nil {
			return 0, debug.WrapError(err, "Error parsing Pricing input file %s", filename)
		}
		debug.Info("Got input: %+v", input)
	*/

	debug.Info(`Completed reading the Pricing input file "%s" :  %d records`, filename, len(allPricingInfo))
	return len(allPricingInfo), nil
}

// hasCachedPricingInfo returns true if we have cached Pricing info loaded from a file
func hasCachedPricingInfo() bool {
	return allPricingInfo != nil
}

// getCachedPricingInfo returns the cached Pricing info for a given Catalog entry (by Catalog ID)
func getCachedPricingInfo(id string) *PricingInfo {
	return allPricingInfo[id]
}

func parseJSONError(err error) string {
	switch err1 := err.(type) {
	case *json.SyntaxError:
		return fmt.Sprintf("json.SyntaxError near offset %d", err1.Offset)
	case *json.UnmarshalTypeError:
		return fmt.Sprintf("json.UnmarshalTypeError near offset %d", err1.Offset)
	default:
		return fmt.Sprintf("%T (offset unknown)", err)
	}
}
