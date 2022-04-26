// osscatpricingloader: a tool to load pricing information from the Global Catalog / Pricing Catalog and
// store it into a set of files, for subsequent use as input in osscatimporter.
// This tool is necessary because loading from the Pricing Catalog can be extremely time-consuming,
// so we need a way to do this incrementally.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"

	"github.com/pkg/profile"
	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
)

// AppVersion is the current version number for the osscatimporter program
const AppVersion = "0.1"

// AppName is the external name of this application
const AppName = "osscatpricingloader"

var versionFlag = flag.Bool("version", false, "Print the version number of this program")
var patternString = flag.String("pattern", "", "A regular expression pattern for the names of all services/components to be updated or \"ALL\" to update all known services")
var logFileName = flag.String("log", AppName+"-log-<PATTERN>-<timestamp>.log", "Path to a file in which to store a log of all the actions and updates performed by this tool")
var outputFileName = flag.String("output", AppName+"-output-<PATTERN>-<timestamp>.json", "Path to a file in which to store all the pricing information collected (in JSON format)")

func usageError(errorMsg string, useFmt bool) {
	defer func() {
		os.Exit(-1)
	}()
	if useFmt && errorMsg != "" {
		fmt.Println(errorMsg)
	} else if !useFmt {
		debug.Critical(errorMsg)
	}
	fmt.Println("*** Aborting")
}

var countProcessed int
var countWithData int

// Options is a pointer to the global options.
var Options *options.Data

func main() {
	var serviceNameMsg string
	var patternDesc string
	var err error

	Options = options.LoadGlobalOptions("", false)

	if debug.IsDebugEnabled(debug.Profiling) {
		debug.Info("CPU profiling enabled")
		defer profile.Start().Stop()
	}

	if *patternString == "" {
		usageError("-pattern flag missing", true)
	}
	if strings.ToLower(*patternString) == "all" {
		*patternString = ".*"
		serviceNameMsg = fmt.Sprintf("all services and components found from all sources (with visibility=%s and above)", Options.VisibilityRestrictions)
		patternDesc = "ALL"
	} else {
		serviceNameMsg = fmt.Sprintf(`services and components matching the pattern "%s" (with visibility=%s and above)`, *patternString, Options.VisibilityRestrictions)
		patternDescRegexp := regexp.MustCompile(`[^a-zA-Z0-9-]`)
		patternDesc = patternDescRegexp.ReplaceAllLiteralString(strings.TrimSpace(*patternString), "_")
	}
	pattern, err := regexp.Compile(*patternString)
	if err != nil {
		fmt.Println(debug.WrapError(err, `Invalid service name pattern: "%s"`, *patternString))
		return
	}

	debug.Info("%s Version: %s\n" /*os.Args[0]*/, AppName, AppVersion)
	if *versionFlag {
		return
	}

	var patternSubstRegexp = regexp.MustCompile("<PATTERN>")

	// Open the log file.
	// This must happen before we start generating any meaningful output
	fullLogFileName := options.GetOutputFileName(*logFileName, options.RunModeRO, "", "log", "log")
	fullLogFileName = patternSubstRegexp.ReplaceAllLiteralString(fullLogFileName, patternDesc)
	logFile, err := os.Create(filepath.Clean(fullLogFileName))
	//logFile, err := os.Create(strings.TrimSpace(fullLogFileName))
	if err != nil {
		panic(debug.WrapError(err, "Cannot create the log file"))
	}
	defer func() {
		logFile.Close()
	}()
	debug.SetLogFile(logFile)

	// Generate final run stats or panic message
	// Must happen after the log file has been set-up and before it gets closed by a deferred function
	defer func() {
		fmt.Println()
		recoveredObj := recover()
		if recoveredObj != nil {
			debug.PanicHandler(recoveredObj)
		} else {
			debug.Audit(fmt.Sprintf("Completed run of %s: %d Catalog entries processed; %d entries with Pricing info\n", AppName, countProcessed, countWithData))
			debug.Audit(debug.SummarizeErrors("        "))
		}
	}()

	// Open the output file
	fullOutputFileName := options.GetOutputFileName(*outputFileName, options.RunModeRO, "", "output", "json")
	fullOutputFileName = patternSubstRegexp.ReplaceAllLiteralString(fullOutputFileName, patternDesc)
	outputFile, err := os.Create(filepath.Clean(fullOutputFileName))
	if err != nil {
		panic(debug.WrapError(err, "Cannot create the output file"))
	}
	io.WriteString(outputFile, fmt.Sprintf("[\n"))
	io.WriteString(outputFile, fmt.Sprintf("   {}")) // Dummy first record, so that we can use a comma to start each new record
	defer func() {
		io.WriteString(outputFile, fmt.Sprintf("\n]\n"))
		// #nosec G307
		outputFile.Close()
		/*
			err := outputFile.Close()
			if err != nil {
				panic(debug.WrapError(err, "Cannot close the output file"))
			}
		*/
	}()

	debug.Audit("Operating on %s\n", serviceNameMsg)

	err = doWork(pattern, outputFile)
	if err != nil {
		debug.Critical("Error loading all source entries for pattern \"%s\": %v", pattern, err)
	}
}

func doWork(pattern *regexp.Regexp, outputFile *os.File) error {
	err := catalog.ListMainCatalogEntries(pattern, func(r *catalogapi.Resource) {
		switch r.Kind {
		case catalogapi.KindPlan, catalogapi.KindFlavor, catalogapi.KindProfile:
			return
		case catalogapi.KindDeployment:
			return
		case catalogapi.KindRegion, catalogapi.KindDatacenter, catalogapi.KindAvailabilityZone, catalogapi.KindPOP, catalogapi.KindLegacyCName, catalogapi.KindLegacyEnvironment, catalogapi.KindSatellite:
			return
		default:
			comparableName := ossmerge.MakeComparableName(r.Name)
			si, _ := ossmerge.LookupService(comparableName, true)
			if !si.HasSourceMainCatalog() {
				si.SourceMainCatalog = *r
			} else {
				si.AdditionalMainCatalog = append(si.AdditionalMainCatalog, r)
			}
			ossmerge.RecordPricingInfo(si, r)
			if Options.Verbose {
				cx := si.GetCatalogExtra(false)
				if cx != nil {
					debug.Info(`Pricing for %s(%s): %v`, r.CatalogPath, si.String(), cx.PartNumbers)
				} else {
					debug.Info(`Pricing for %s(%s): (no CatalogExtra)`, r.CatalogPath, si.String())
				}
			}
			countProcessed++
		}
	})
	if err != nil {
		return (debug.WrapError(err, "Error listing Main entries from Global Catalog"))
	}

	// Generate the output
	debug.Info("Generating output for %d entries", countProcessed)
	err = ossmerge.ListAllServices(pattern, func(si *ossmerge.ServiceInfo) {
		if !si.HasSourceMainCatalog() {
			panic(fmt.Sprintf("ServiceInfo(%s) has not SourceMainCatalog", si.String()))
		}
		r := si.GetSourceMainCatalog()
		if len(si.AdditionalMainCatalog) > 0 {
			debug.Warning("Ignoring %d additional Catalog entries for %s", len(si.AdditionalMainCatalog), r.String())
		}
		cex := si.GetCatalogExtra(false)
		if cex != nil {
			output := ossmerge.PricingInfo{
				CatalogID:   r.ID,
				CatalogName: r.Name,
				CatalogKind: r.Kind,
				CatalogPath: r.CatalogPath,
				PartNumbers: cex.PartNumbers.Slice(),
				Issues:      si.OSSValidation.Issues,
			}
			io.WriteString(outputFile, ",\n")
			io.WriteString(outputFile, output.JSON())
			if !ossrecord.IsProductInfoNone(output.PartNumbers) {
				countWithData++
			} else if len(output.PartNumbers) == 0 {
				panic(fmt.Sprintf("PartNumbers is empty for %s  %s", si.String(), r.String()))
			}
		}
	})
	if err != nil {
		return (debug.WrapError(err, "Error processing loaded entries"))
	}

	return nil
}
