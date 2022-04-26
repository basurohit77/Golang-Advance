// osscatpublisher: tool to copy/publish OSS records from the Staging instance of the Global Catalog
// (which contains the most recently updated data)
// to the Production instance (which contains the stable data)
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	utils "github.ibm.com/cloud-sre/osscatalog/cmd/utils"
	"github.ibm.com/cloud-sre/osscatalog/cos"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/stats"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// AppVersion is the current version number for the osscatpublisher program
const AppVersion = "0.1"

// AppName is the external name of this application
const AppName = "osscatpublisher"

var versionFlag = flag.Bool("version", false, "Print the version number of this program")
var roFlag = flag.Bool("ro", false, "Execute in read-only mode (make no changes to Catalog OSS entries)")
var rwFlag = flag.Bool("rw", false, "Execute in read-write mode (create/update/archive Catalog OSS entries as appropriate)")
var interactiveFlag = flag.Bool("interactive", false, "Execute in interactive mode (review proposed update, then prompt user, then create/update/archive Catalog OSS entries as appropriate)")
var slackChannel = flag.String("slack-channel", "", "Send notification to a slack channel.")
var slackResponder = flag.String("slack-responder", "", "First responder in slack channel.")
var forceFlag = flag.Bool("force", false, "Force rewriting existing Catalog OSS entries even if they are not modified")
var serviceName = flag.String("service", "", "Name of one single service or component to be updated")
var cosEndpoint = flag.String("cosendpoint", "", "Endpoints are used hand in hand with your credentials (i.e. keys, CRN, bucket name) to tell your service where to look for this bucket. Depending on where your service or applications is located you will want to use one of the below endpoint types. For example : s3.us-east.cloud-object-storage.appdomain.cloud")
var cosBucketName = flag.String("cosbucketname", "", "Name of COS bucket. For example : rmc-data")
var cosEmergBucketName = flag.String("cosemergencybucketname", "", "Name of Emergency COS bucket. For example : oss-rmc-emergency")
var cosBucketLink = flag.String("cosbucketlink", "", "A href link of COS bucket. For example : https://cloud.ibm.com/objectstorage/crn%3Av1%3Abluemix%3Apublic%3Acloud-object-storage%3Aglobal%3Aa%2F0bb4d59c58f057ca240dd82f9bc42eb3%3A32d69b06-6a20-4418-88cd-8605d38ff183%3A%3A?bucket=rmc-data&bucketRegion=us-east&endpoint=s3.us-east.cloud-object-storage.appdomain.cloud&paneId=bucket_overview")
var patternString = flag.String("pattern", "", "A regular expression pattern for the names of all services/components, segments, tribes and environments to be updated or \"ALL\" to update all known entries")
var outputFileName = flag.String("output", AppName+"-<mode>-output-<timestamp>.json", "Path to a file in which to store all the new/updated OSS records (in JSON format), regardless of whether they actually got written to the Catalog itself")
var logFileName = flag.String("log", AppName+"-<mode>-log-<timestamp>.log", "Path to a file in which to store a log of all the actions and updates performed by this tool")
var inputFileName = flag.String("input", "", "Path to an input file to use instead of reading source entries from Staging Catalog -- normally the output file from an earlier run to osscatimporter")
var excludeEnvironments = flag.Bool("no-environments", false, "Do not publish OSS records for Environments")
var stagingOnly = flag.Bool("staging", false, "Publish to the Staging environment instead of Production (requires the -input parameter)")
var productionOnly = flag.Bool("production", false, "Publish to the Production environment and not Staging (default unless the -input parameter is specified)")
var incremental = flag.Bool("incremental", false, "Incremental mode: load new entries from the -input file but do not delete any other entries")

var outputFile *os.File
var logFile *os.File
var inputFile *os.File

// store all actions will be take later
var catalogActions []*preparedCatalogAction

type preparedCatalogAction struct {
	target          ossrecord.OSSEntry
	prior           ossrecord.OSSEntry
	diffs           *compare.Output
	action          stats.Action
	updateViolation bool
}

// Options is a pointer to the global options.
var Options *options.Data

type entry struct {
	id      ossrecord.OSSEntryID
	typ     reflect.Type
	sortKey string
	source  ossrecord.OSSEntry
	dest    ossrecord.OSSEntry
}

var allEntries map[ossrecord.OSSEntryID]*entry

var typeEnvironment = reflect.TypeOf(&ossrecord.OSSEnvironment{})
var typeEnvironmentExtended = reflect.TypeOf(&ossrecordextended.OSSEnvironmentExtended{})
var typeSegment = reflect.TypeOf(&ossrecord.OSSSegment{})
var typeSegmentExtended = reflect.TypeOf(&ossrecordextended.OSSSegmentExtended{})
var typeTribe = reflect.TypeOf(&ossrecord.OSSTribe{})
var typeTribeExtended = reflect.TypeOf(&ossrecordextended.OSSTribeExtended{})
var typeService = reflect.TypeOf(&ossrecord.OSSService{})
var typeServiceExtended = reflect.TypeOf(&ossrecordextended.OSSServiceExtended{})
var typeOSSResourceClassification = reflect.TypeOf(&ossrecord.OSSResourceClassification{})

var serviceNameMsg string
var sourceName string
var destName string
var runMode options.RunMode
var stagingTag string
var incrementalTag string
var fullLogFileName string

func usageError(errorMsg string, useFmt bool) {
	defer func() {
		os.Exit(-1)
	}()
	if useFmt {
		fmt.Println(errorMsg)
	} else {
		debug.Critical(errorMsg)
	}
	fmt.Println()
	fmt.Println("Usage:")
	flag.PrintDefaults()
	utils.PostSlackMessage(AppName, *slackChannel, "Usage Error", errorMsg, utils.ErrorType, "")
}
func main() {
	var err error
	Options = options.LoadGlobalOptions("", false)
	defer func() {
		utils.ExitHandler(AppName, *slackChannel, *slackResponder, debug.GetPanicMessage())
	}()
	if os.Getenv("RUNNING_IN_KUBE") == "yes" && (*slackChannel == "" || *slackResponder == "" || *cosEndpoint == "" || *cosBucketName == "" || *cosEmergBucketName == "") {
		usageError("-slack-channel, -slack-responder, -cosendpoint, -cosbucketname, and -cosemergencybucketname are required parameters when "+AppName+" running in a kube cluster.", true)
	}
	if *slackChannel != "" {
		err = utils.PostSlackMessage(AppName, *slackChannel, "Starting Run "+AppName, strings.Join(os.Args, " "), utils.InfoType, "")
		if err != nil {
			usageError("invalid slack configuration.", true)
		}
	}

	// Figure out the run mode
	if *interactiveFlag {
		runMode = options.RunModeInteractive
		if *rwFlag || *roFlag {
			runMode = options.RunModeRO
			usageError("-interactive flag specified together with -ro or -rw flag", false)
		}
	} else if *rwFlag {
		runMode = options.RunModeRW
		if *roFlag {
			runMode = options.RunModeRO
			usageError("-ro and -rw flags both specified", false)
		}
	} else {
		runMode = options.RunModeRO
	}

	// Make sure COS service is available. Only upload files to COS if run in ready-write mode
	if *cosEndpoint != "" {
		// test regular
		isPass, err := cos.TestConnection(*cosEndpoint, *cosBucketName)
		if !isPass || err != nil {
			detailError := ""
			if err != nil {
				detailError = err.Error()
			}
			usageError(fmt.Sprintf("Specified COS service is not available(%s %s). Error is : %s", *cosEndpoint, *cosBucketName, detailError), true)
		} else {
			// do not download

			//_, err = cos.Download(*cosEndpoint, *cosBucketName, "run-"+AppName+".txt")
			//if err != nil {
			//	usageError("Skip run "+AppName+", for run-"+AppName+".txt file was not found in cos "+*cosBucketName, true)
			//}
		}
		// test emerg
		isPassEmerg, errEmerg := cos.TestConnection(*cosEndpoint, *cosEmergBucketName)
		if !isPassEmerg || errEmerg != nil {
			detailError := ""
			if errEmerg != nil {
				detailError = errEmerg.Error()
			}
			usageError(fmt.Sprintf("Specified COS service is not available(%s %s). Error is : %s", *cosEndpoint, *cosEmergBucketName, detailError), true)
		} else {
			_, err = cos.Download(*cosEndpoint, *cosEmergBucketName, "run-"+AppName+".txt")
			if err != nil {
				usageError("Skip run "+AppName+", for run-"+AppName+".txt file was not found in cos "+*cosEmergBucketName, true)
			}
		}
	}
	var pattern *regexp.Regexp
	if *serviceName != "" && *patternString != "" {
		usageError("-service and -pattern flags both specified", true)
	}
	if *stagingOnly && *productionOnly {
		usageError("-staging and -production flags both specified", false)
	}
	if *stagingOnly {
		stagingTag = "-staging"
		if *inputFileName == "" {
			usageError("-staging requires the -input parameter", true)
		}
		sourceName = "Inputfile"
		destName = "Staging Catalog (as destination)"
		if *serviceName != "" {
			usageError("-staging and -service flags are incompatible", true)
		}
	} else {
		if *inputFileName == "" {
			sourceName = "Staging Catalog"
		} else {
			sourceName = "Inputfile"
		}
		destName = "Production Catalog"
	}
	if *incremental {
		incrementalTag = "-incremental"
		if *inputFileName == "" {
			usageError("-incremental requires the -input parameter", true)
		}
		if !*stagingOnly && !*productionOnly {
			usageError("-incremental requires the -staging or -production flag to be explicitly specified", false)
		}
	}
	switch {
	case *serviceName != "":
		serviceNameMsg = fmt.Sprintf("service \"%s\" only", *serviceName)
		pattern = nil // Just to make sure but should already be zero
	case *patternString != "":
		if strings.ToLower(*patternString) == "all" {
			*patternString = ".*"
			serviceNameMsg = fmt.Sprintf("all OSS entries (with visibility=%s and above)", Options.VisibilityRestrictions)
		} else {
			serviceNameMsg = fmt.Sprintf(`OSS entries matching the pattern "%s" (with visibility=%s and above)`, *patternString, Options.VisibilityRestrictions)
		}
		pattern, err = regexp.Compile(*patternString)
		if err != nil {
			fmt.Println(debug.WrapError(err, `Invalid service name pattern: "%s"`, *patternString))
			return
		}
		*serviceName = "" // Just to make sure but should already be zero
	default:
		usageError("Must specify one of -service or -pattern flags", true)
	}
	if *excludeEnvironments {
		debug.Info("Omitting OSS environments in this run")
	}

	// Open the input file.
	if *inputFileName != "" {
		inputFile, err = os.Open(*inputFileName)
		if err != nil {
			panic(debug.WrapError(err, "Cannot open the input file"))
		}
		defer inputFile.Close() // #nosec G307
	}

	// Open the log file.
	// This must happen before we start generating any meaningful output
	fullLogFileName = options.GetOutputFileName(*logFileName, runMode, stagingTag+incrementalTag, "log", "log")
	logFile, err = os.Create(filepath.Clean(fullLogFileName))
	if err != nil {
		panic(debug.WrapError(err, "Cannot create the log file"))
	}
	defer func() {
		logFile.Close()
		uploadFileToCOS(fullLogFileName, cos.CONTENT_TYPE_TEXT)
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
			var globalStatsReportOutput string
			globalStatsReportOutput = stats.GetGlobalActualStats().Report("        ")
			debug.Audit(fmt.Sprintf("Completed run of "+AppName+" (%s mode): \n%s", runMode.String()+incrementalTag, globalStatsReportOutput))
			utils.PostSlackMessage(AppName, *slackChannel, fmt.Sprintf("Completed run of %s (%s mode)", AppName, runMode.String()+incrementalTag), globalStatsReportOutput, utils.InfoType, "")
			debug.Audit(debug.SummarizeErrors("        "))
		}
	}()

	if runMode == options.RunModeRW || runMode == options.RunModeInteractive {
		options.BootstrapLogDNA(AppName+`-`+runMode.ShortString(), true, 0)
	}

	debug.Info("%s Version: %s\n" /*os.Args[0]*/, AppName, AppVersion)
	if *versionFlag {
		return
	}

	if runMode == options.RunModeRW || runMode == options.RunModeInteractive {
		if *stagingOnly {
			options.SetProductionWriteEnabled(false)
			debug.Info("Operating in %s mode on %s -- Destination = Staging URL=%s\n", runMode, serviceNameMsg, catalog.GetOSSStagingURL(true))
		} else {
			options.SetProductionWriteEnabled(true)
			debug.Info("Operating in %s mode on %s -- Destination = PRODUCTION URL=%s\n", runMode, serviceNameMsg, catalog.GetOSSProductionURL(true))
		}
	} else {
		options.SetProductionWriteEnabled(false)
		if *stagingOnly {
			debug.Info("Operating in %s mode on %s -- Destination = Staging URL=%s\n", runMode, serviceNameMsg, catalog.GetOSSStagingURL(false))
		} else {
			debug.Info("Operating in %s mode on %s -- Destination = Production URL=%s\n", runMode, serviceNameMsg, catalog.GetOSSProductionURL(false))
		}
	}

	// Check OSS entry ownership
	if options.IsCheckOwnerSpecified() {
		debug.Info(`Using -check-owner=%v (explicitly specified on command line)`, Options.CheckOwner)
	} else {
		Options.CheckOwner = true
		debug.Info(`Forcing -check-owner=%v`, Options.CheckOwner)
	}

	// Open the output file
	fullOutputFileName := options.GetOutputFileName(*outputFileName, runMode, stagingTag+incrementalTag, "output", "json")
	outputFile, err = os.Create(filepath.Clean(fullOutputFileName))
	if err != nil {
		panic(debug.WrapError(err, "Cannot create the output file"))
	}
	io.WriteString(outputFile, fmt.Sprintf("[\n"))
	io.WriteString(outputFile, fmt.Sprintf("  {}")) // Dummy first record, so that we can use a comma to start each new record
	/* #nosec */
	defer func() {
		io.WriteString(outputFile, fmt.Sprintf("]\n"))
		// #nosec G307
		outputFile.Close()
		/*
			err := outputFile.Close()
			if err != nil {
				panic(debug.WrapError(err, "Cannot close the output file"))
			}
		*/
		uploadFileToCOS(fullOutputFileName, cos.CONTENT_TYPE_TEXT)
	}()

	if pattern != nil {
		allEntries = make(map[ossrecord.OSSEntryID]*entry)

		if inputFile != nil {
			numEntries, err := parseJSONInput(inputFile, pattern, oneSourceEntry)
			if err != nil {
				panic(fmt.Sprintf("Error loading all OSS entries from input file \"%s\" for pattern \"%s\": %v", *inputFileName, pattern, err))
			}
			debug.Info("Read %d OSS entries from input file %s", numEntries, *inputFileName)
		} else {
			debug.Info("Loading OSS entries from Staging Global Catalog")
			// Include services domain overrides as the full record needs to be written to production:
			err := catalog.ListOSSEntries(pattern, catalog.IncludeServices|catalog.IncludeTribes|catalog.IncludeEnvironments|catalog.IncludeServicesDomainOverrides|catalog.IncludeOSSResourceClassification, oneSourceEntry)
			if err != nil {
				panic(fmt.Sprintf("Error loading all OSS entries from Staging Catalog for pattern \"%s\": %v", pattern, err))
			}
		}

		if *stagingOnly {
			debug.Info("Loading OSS entries from Staging Catalog (used as destination)")
			// Include services domain overrides to allow for diff against source:
			err = catalog.ListOSSEntries(pattern, catalog.IncludeServices|catalog.IncludeTribes|
				catalog.IncludeEnvironments|catalog.IncludeOSSMergeControl|catalog.IncludeOSSValidation|catalog.IncludeServicesDomainOverrides|catalog.IncludeOSSResourceClassification, func(r ossrecord.OSSEntry) {
				id := r.GetOSSEntryID()
				e, found := allEntries[id]
				if !found {
					e = &entry{}
					allEntries[id] = e
				}
				if e.dest != nil {
					panic(fmt.Sprintf(`Duplicate entry ID "%s" in Staging Catalog (used as destination)`, id))
				}
				e.dest = r
			})
			if err != nil {
				panic(fmt.Sprintf("Error loading all OSS entries from Staging Catalog (used as destination) for pattern \"%s\": %v", pattern, err))
			}
		} else {
			debug.Info("Loading OSS entries from Production Catalog")
			// Include services domain overrides to allow for diff against source:
			err = catalog.ListOSSEntriesProduction(pattern, catalog.IncludeServices|catalog.IncludeTribes|catalog.IncludeEnvironments|catalog.IncludeServicesDomainOverrides|catalog.IncludeOSSResourceClassification, func(r ossrecord.OSSEntry) {
				id := r.GetOSSEntryID()
				e, found := allEntries[id]
				if !found {
					e = &entry{}
					allEntries[id] = e
				}
				if e.dest != nil {
					panic(fmt.Sprintf(`Duplicate entry ID "%s" in Production Catalog`, id))
				}
				e.dest = r
			})
			if err != nil {
				panic(fmt.Sprintf("Error loading all OSS entries from Production Catalog for pattern \"%s\": %v", pattern, err))
			}
		}
		sortedEntries := make([]*entry, 0, len(allEntries))
		for id, e := range allEntries {
			e.id = id
			switch {
			case e.source != nil && e.dest != nil:
				styp := reflect.TypeOf(e.source)
				ptyp := reflect.TypeOf(e.dest)
				if styp != ptyp {
					panic(fmt.Sprintf(`Conflicting types for entry id="%s": %s=%v  %s=%v`, id, sourceName, ptyp, destName, ptyp))
				}
				e.typ = ptyp
			case e.source != nil && e.dest == nil:
				e.typ = reflect.TypeOf(e.source)
			case e.source == nil && e.dest != nil:
				e.typ = reflect.TypeOf(e.dest)
			case e.source == nil && e.dest == nil:
				panic(fmt.Sprintf(`Entry id="%s" contains no data from %s or %s`, id, sourceName, destName))
			}
			switch e.typ {
			case typeOSSResourceClassification:
				e.sortKey = "0." + string(e.id)
			case typeSegment, typeSegmentExtended:
				e.sortKey = "1." + string(e.id)
			case typeTribe, typeTribeExtended:
				e.sortKey = "2." + string(e.id)
			case typeEnvironment, typeEnvironmentExtended:
				e.sortKey = "3." + string(e.id)
			case typeService, typeServiceExtended:
				e.sortKey = "4." + string(e.id)
			default:
				panic(fmt.Sprintf("Unexpected entry type: %v", e.typ))
			}
			sortedEntries = append(sortedEntries, e)
		}
		sort.Slice(sortedEntries, func(i, j int) bool {
			return sortedEntries[i].sortKey < sortedEntries[j].sortKey
		})
		debug.Info("\n\n================= Review/Prepare phase: showing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())
		for _, e := range sortedEntries {
			prepareCatalogAction(e)
		}
	} else {
		e := &entry{}
		var err error
		// Include services domain overrides as the full record needs to be written to production:
		e.source, err = catalog.ReadOSSServiceWithOptions(ossrecord.CRNServiceName(*serviceName), catalog.IncludeServicesDomainOverrides)
		if err != nil {
			debug.Info(`Entry "%s" not found in Staging Catalog`, *serviceName)
			e.source = nil
		}
		// Include services domain overrides to allow for diff against source:
		e.dest, err = catalog.ReadOSSServiceProductionWithOptions(ossrecord.CRNServiceName(*serviceName), catalog.IncludeServicesDomainOverrides)
		if err != nil {
			debug.Info(`Entry "%s" not found in Production Catalog`, *serviceName)
			e.dest = nil
		}
		switch {
		case e.source != nil && e.dest != nil:
			sid := e.source.GetOSSEntryID()
			pid := e.dest.GetOSSEntryID()
			if sid != pid {
				panic(fmt.Sprintf(`Mismatched IDs when fetching entries for service "%s": %s ID="%s" / %s ID="%s"`, *serviceName, sourceName, sid, destName, pid))
			}
			e.id = pid
			styp := reflect.TypeOf(e.source)
			ptyp := reflect.TypeOf(e.dest)
			if styp != ptyp {
				panic(fmt.Sprintf(`Conflicting types for entry "%s", id="%s" in Catalog: %s=%v  %s=%v`, *serviceName, e.id, sourceName, ptyp, destName, ptyp))
			}
		case e.source != nil && e.dest == nil:
			e.id = e.source.GetOSSEntryID()
			e.typ = reflect.TypeOf(e.source)
		case e.source == nil && e.dest != nil:
			e.id = e.dest.GetOSSEntryID()
			e.typ = reflect.TypeOf(e.dest)
		case e.source == nil && e.dest == nil:
			panic(fmt.Sprintf(`Entry "%s" not found in %s or %s`, *serviceName, sourceName, destName))
		}
		debug.Info("\n\n================= Review/Prepare phase: showing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())
		prepareCatalogAction(e)
	}
	performCatalogActions()
	rest.PrintTimings()
}

func prepareCatalogAction(e *entry) {
	var action stats.Action
	var diffs compare.Output
	var target, prior ossrecord.OSSEntry
	switch {
	case e.source != nil && e.dest != nil && e.source.GetOSSTags().Contains(osstags.OSSDelete):
		action = stats.ActionDelete
		target = e.dest
		prior = e.dest
	case e.source != nil && e.source.GetOSSTags().Contains(osstags.OSSStaging) && !*stagingOnly:
		action = stats.ActionIgnore(fmt.Sprintf(`Entry is marked ""%s": "%s"`, osstags.OSSStaging, e.source.String()))
		target = e.source
	case inputFile == nil && e.source != nil && e.source.GetOSSOnboardingPhase() != "" && e.source.GetOSSOnboardingPhase() != ossrecord.PRODUCTION:
		action = stats.ActionIgnore(fmt.Sprintf(`Entry is managed through RMC and the phase is not "Production" (OSSOnboardingPhase=%s): "%s"`, e.source.GetOSSOnboardingPhase(), e.source.String()))
		target = e.source
	case e.source != nil && e.source.GetOSSTags().Contains(osstags.OSSTest) && !*stagingOnly:
		action = stats.ActionIgnore(fmt.Sprintf(`test entry: "%s"`, e.source.String()))
		target = e.source
	case e.source != nil && e.dest != nil:
		diffs = compare.Output{}
		if !*stagingOnly {
			e.source.CleanEntryForCompare()
			e.dest.CleanEntryForCompare()
		}
		compare.DeepCompare("before", e.dest, "after", e.source, &diffs)
		numDiffs := diffs.NumDiffs()
		if runMode == options.RunModeRW {
			if e.source.GetOSSTags().Contains(osstags.OSSLock) && !*stagingOnly {
				action = stats.ActionLocked
			} else if *forceFlag {
				action = stats.ActionUpdate
			} else if numDiffs == 0 {
				action = stats.ActionNotModified
			} else {
				action = stats.ActionUpdate
			}
		} else {
			if e.source.GetOSSTags().Contains(osstags.OSSLock) && !*stagingOnly {
				action = stats.ActionLocked
			} else if numDiffs == 0 {
				action = stats.ActionNotModified
			} else {
				action = stats.ActionUpdate
			}
		}
		target = e.source
		prior = e.dest
	case e.source != nil && e.dest == nil:
		if e.source.GetOSSTags().Contains(osstags.OSSLock) && !*stagingOnly {
			action = stats.ActionLocked
		} else {
			action = stats.ActionCreate
		}
		target = e.source
	case e.source == nil && e.dest != nil:
		if *incremental {
			action = stats.ActionIgnore(fmt.Sprintf(`incremental mode+not present in input file: "%s"`, e.dest.String()))
		} else {
			action = stats.ActionDelete
		}
		target = e.dest
		prior = e.dest
	case e.source == nil && e.dest == nil:
		debug.PrintError(`Both %s and %s records are empty for id="%s"`, sourceName, destName, e.id)
		action = stats.ActionError
		return
	}

	var updateViolation bool
	if !target.IsUpdatable() {
		action = stats.ActionCatalogNative
		if action != stats.ActionNotModified && !action.IsActionIgnore() {
			updateViolation = true
		}
	}

	// Write an entry in the logFile and stdout
	var logString = "*** NO LOG ***"
	if logFile != nil || Options.Verbose {
		logBuffer := new(strings.Builder)
		logBuffer.WriteString("\n")
		logBuffer.WriteString("*** ")
		logBuffer.WriteString(action.Message(runMode, *forceFlag))
		logBuffer.WriteString(" --> ")
		logBuffer.WriteString(target.Header())
		diffString := diffs.StringWithPrefix("  UPDATED:")
		if len(diffString) > 0 {
			logBuffer.WriteString(fmt.Sprintf("---- BEGIN DIFFS %s\n", diffs.Summary()))
			logBuffer.WriteString(diffString)
			logBuffer.WriteString("---- END   DIFFS \n")
		}
		logBuffer.WriteString("\n")
		logString = logBuffer.String()
		if logFile != nil {
			debug.PlainLogEntry(debug.LevelINFO, target.String(), logString)
		}
	}
	if Options.Verbose {
		fmt.Print(logString)
	} else if action != stats.ActionNotModified {
		fmt.Println()
		fmt.Print("*** ", action.Message(runMode, *forceFlag))
		fmt.Print(" --> ")
		fmt.Print(target.Header())
		fmt.Println()
	}

	if updateViolation {
		debug.PrintError("Found updatable differences in a non-updatable entry: %s", target.String())
	}

	// Update the counts
	stats.GetGlobalPrepareStats().RecordAction(action, runMode, *forceFlag, target, prior)
	catalogActions = append(catalogActions, &preparedCatalogAction{target: target, prior: prior, diffs: &diffs, action: action, updateViolation: updateViolation})
}

func performCatalogActions() {
	abortMessage := stats.GetGlobalPrepareStats().CheckForAbort()
	if len(abortMessage) > 0 {
		switch runMode {
		case options.RunModeRW:
			debug.Critical("ABORTING READ-WRITE run and switching to read-only mode: \n%s", abortMessage)
			runMode = options.RunModeRO
			options.SetProductionWriteEnabled(false)
		case options.RunModeRO:
			debug.Critical("Proceeding with read-only run despite issues that would trigger an abort in READ-WRITE mode: \n%s", abortMessage)
			options.SetProductionWriteEnabled(false)
		case options.RunModeInteractive:
			debug.Critical("Prompting user for whether to proceed despite issues that would trigger an abort in READ-WRITE mode: \n%s", abortMessage)
		default:
			panic(fmt.Sprintf(`Unknown runMode: "%v"`, runMode))
		}
	}

	if runMode == options.RunModeInteractive {
		globalStatsReportOutput := stats.GetGlobalPrepareStats().Report("        ")
		debug.Info(fmt.Sprintf("\n\nRunning %s in %s mode - review summary for evaluation: \n%s", AppName, runMode.String()+incrementalTag, globalStatsReportOutput))
		debug.Info(debug.SummarizeErrors("        "))
		err := logFile.Sync()
		if err != nil {
			debug.PrintError("Error syncing log file: %v", err)
		}
		fmt.Println()
		runMode = options.RunModeRO
		//First upload the logFile in progress
		uploadFileToCOS(fullLogFileName, cos.CONTENT_TYPE_TEXT)
	loop:
		for {
			var input string
			if os.Getenv("RUNNING_IN_KUBE") == "yes" {
				// Send slack notification for prompt
				podName := os.Getenv("HOSTNAME")
				containerName := os.Getenv("CONTAINER_NAME")
				instructionMessage := fmt.Sprintf("Please use command `kubectl oss attach %s --namespace api -c %s -i -t` to access the container. Enter `continue` to proceed with write, `readonly` to continue with read only mode, or `stop` to exit the program. ", podName, containerName)
				slackMsgTitle := "Prompting for user input to either proceed with writing the output to Catalog OSS entries or to abort"
				slackMsgBody := fmt.Sprintf("Prompting user for whether to proceed despite issues that would trigger an abort in READ-WRITE mode: \n%s\n%s", abortMessage, instructionMessage)
				utils.PostSlackMessage(AppName, *slackChannel, slackMsgTitle, slackMsgBody, utils.ErrorType, *slackResponder)
			}
			fmt.Print(`Type "continue" to proceed with writing the output to Catalog OSS entries, "readonly" to continue with read only mode, or "stop" to abort: `)
			fmt.Scanln(&input)
			input = strings.TrimSpace(input)
			switch input {
			case "continue":
				runMode = options.RunModeRW
				break loop
			case "stop":
				options.SetProductionWriteEnabled(false)
				os.Exit(-1)
			case "readonly":
				runMode = options.RunModeRO
				options.SetProductionWriteEnabled(false)
				break loop
			}
		}
	}

	// count := len(catalogActions)
	debug.Info("\n\n================= Output/Commit phase: finalizing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())
	for _ /* i */, catalogAction := range catalogActions {
		//fmt.Println("do " + strconv.Itoa(i) + "/" + strconv.Itoa(count) + "...")
		performCatalogAction(catalogAction)
	}
}

func performCatalogAction(preparedCatalogAction *preparedCatalogAction) {
	var action stats.Action = preparedCatalogAction.action
	var target ossrecord.OSSEntry = preparedCatalogAction.target
	var prior ossrecord.OSSEntry = preparedCatalogAction.prior

	if action == stats.ActionUpdate {
		if runMode == options.RunModeRW {
			var err error
			if *stagingOnly {
				err = catalog.UpdateOSSEntry(target, catalog.IncludeAll)
			} else {
				err = catalog.UpdateOSSEntryProduction(target, catalog.IncludeNone)
			}
			if err != nil {
				debug.PrintError("Error updating OSS Catalog entry for \"%s\": %v", target.String(), err)
				action = stats.ActionError
			}
		}
	} else if action == stats.ActionCreate {
		if runMode == options.RunModeRW {
			var err error
			if *stagingOnly {
				err = catalog.CreateOSSEntry(target, catalog.IncludeAll)
			} else {
				err = catalog.CreateOSSEntryProduction(target, catalog.IncludeNone)
			}
			if err != nil {
				httpErr, isHTTPError := err.(rest.HTTPError)
				if action == stats.ActionCreate && *forceFlag && isHTTPError && httpErr.GetHTTPStatusCode() == http.StatusConflict {
					var err2 error
					if *stagingOnly {
						err2 = catalog.UpdateOSSEntry(target, catalog.IncludeAll)
					} else {
						err2 = catalog.UpdateOSSEntryProduction(target, catalog.IncludeNone)
					}
					if err2 != nil {
						debug.PrintError("Error creating OSS Catalog entry for \"%s\": %v --- and attempt at update (in -force mode) also failed: %v", target.String(), err, err2)
						action = stats.ActionError
					} else {
						debug.Warning("Successfully updated OSS Catalog entry for \"%s\" in -force mode after initial create attempt failed --- original error: %v", target.String(), err)
					}
				} else {
					debug.PrintError("Error creating OSS Catalog entry for \"%s\": %v", target.String(), err)
					action = stats.ActionError
				}
			}
		}
	} else if action == stats.ActionDelete {
		if runMode == options.RunModeRW {
			var err error
			if *stagingOnly {
				err = catalog.DeleteOSSEntry(target)
			} else {
				err = catalog.DeleteOSSEntryProduction(target)
			}
			if err != nil {
				debug.PrintError("Error deleting OSS Catalog entry for \"%s\": %v", target.String(), err)
				action = stats.ActionError
			}
		}
	} else if action == stats.ActionLocked {
		// ignore
	} else if action == stats.ActionCatalogNative {
		// ignore
	} else if action.IsActionIgnore() {
		// ignore
	}

	// Update the counts
	stats.GetGlobalActualStats().RecordAction(action, runMode, *forceFlag, target, prior)

	// Write an entry in the logFile and stdout
	// We do this *after* the update, so that any errors will show-up before the update message
	logBuffer := new(strings.Builder)
	logBuffer.WriteString("*** ")
	logBuffer.WriteString(action.Message(runMode, *forceFlag))
	logBuffer.WriteString(" --> ")
	logBuffer.WriteString(target.Header())
	logString := logBuffer.String()
	if logFile != nil {
		debug.PlainLogEntry(debug.LevelINFO, target.String(), logString)
	}
	fmt.Print(logString)

	if action == stats.ActionError || action == stats.ActionDelete || action.IsActionIgnore() {
		return
	}

	// Write an entry in the outputFile
	if outputFile != nil {
		io.WriteString(outputFile, ",\n")
		io.WriteString(outputFile, target.JSON())
	}
}

func oneSourceEntry(r ossrecord.OSSEntry) {
	id := r.GetOSSEntryID()
	// Check for test entries.
	// We cannot solely rely on the oss_test tag here, because we might see some very recent entries
	// that have never been processed by osscatimporter yet
	switch r1 := r.(type) {
	case *ossrecord.OSSService:
		osstags.CheckOSSTestTag(&r1.ReferenceDisplayName, &r1.GeneralInfo.OSSTags)
	case *ossrecordextended.OSSServiceExtended:
		osstags.CheckOSSTestTag(&r1.OSSService.ReferenceDisplayName, &r1.OSSService.GeneralInfo.OSSTags)
	case *ossrecord.OSSEnvironment:
		if *excludeEnvironments {
			debug.Debug(debug.Fine, "Skipping environment entry: %s", string(id))
			return
		}
		osstags.CheckOSSTestTag(&r1.DisplayName, &r1.OSSTags)
	case *ossrecordextended.OSSEnvironmentExtended:
		if *excludeEnvironments {
			debug.Debug(debug.Fine, "Skipping environment entry: %s", string(id))
			return
		}
		osstags.CheckOSSTestTag(&r1.OSSEnvironment.DisplayName, &r1.OSSEnvironment.OSSTags)
	case *ossrecord.OSSSegment:
		osstags.CheckOSSTestTag(&r1.DisplayName, &r1.OSSTags)
	case *ossrecordextended.OSSSegmentExtended:
		osstags.CheckOSSTestTag(&r1.OSSSegment.DisplayName, &r1.OSSSegment.OSSTags)
	case *ossrecord.OSSTribe:
		osstags.CheckOSSTestTag(&r1.DisplayName, &r1.OSSTags)
	case *ossrecordextended.OSSTribeExtended:
		osstags.CheckOSSTestTag(&r1.OSSTribe.DisplayName, &r1.OSSTribe.OSSTags)
	case *ossrecord.OSSResourceClassification:
		// Nothing to check for the singleton OSSResourceClassification object
	default:
		panic(fmt.Sprintf(`Unknown entry type: %v`, r))
	}
	debug.Debug(debug.Fine, "Processing entry: %s", string(id))
	e, found := allEntries[id]
	if !found {
		e = &entry{}
		allEntries[id] = e
	}
	if e.source != nil {
		panic(fmt.Sprintf(`Duplicate entry ID "%s" in %s`, id, sourceName))
	}
	e.source = r
}

func uploadFileToCOS(fullFileName, contentType string) {
	if *cosEndpoint != "" {
		isUploaded, err := cos.Upload(*cosEndpoint, *cosBucketName, fullFileName, contentType)
		if err != nil || !isUploaded {
			errTitle := "Failed to upload file(" + fullFileName + ") to COS " + *cosBucketName
			fmt.Println(errTitle, err)
			errBody := ""
			if os.Getenv("RUNNING_IN_KUBE") == "yes" {
				errBody += "Pod name may be " + os.Getenv("HOSTNAME") + ".\n"
			}
			if err != nil {
				errBody += err.Error()
			}
			utils.PostSlackMessage(AppName, *slackChannel, errTitle, errBody, utils.ErrorType, *slackResponder)
		} else {
			messageBody := ""
			if *cosBucketLink != "" {
				messageBody = "<" + *cosBucketLink + "&prefix=" + fullFileName + "|Open it in the browser.>"
			}
			utils.PostSlackMessage(AppName, *slackChannel, fullFileName+" was uploaded", messageBody, utils.InfoType, "")
		}
	}
}
