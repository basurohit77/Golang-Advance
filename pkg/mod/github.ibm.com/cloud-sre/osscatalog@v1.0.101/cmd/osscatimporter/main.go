// osscatimporter: main tool to import OSS data from all available sources, validate it, and publish OSS records
// to the Staging instance of the Global Catalog (which contains the most recently updated data)
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/profile"
	utils "github.ibm.com/cloud-sre/osscatalog/cmd/utils"
	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/cos"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/supportcenter"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/legacy"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossbackup"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/ossreportsregistry"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/partsinput"
	"github.ibm.com/cloud-sre/osscatalog/stats"

	"io/ioutil"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/servicenow"
)

// AppVersion is the current version number for the osscatimporter program
const AppVersion = "0.1"

// AppName is the external name of this application
const AppName = "osscatimporter"

var versionFlag = flag.Bool("version", false, "Print the version number of this program")
var roFlag = flag.Bool("ro", false, "Execute in read-only mode (make no changes to Catalog OSS entries)")
var rwFlag = flag.Bool("rw", false, "Execute in read-write mode (create/update/archive Catalog OSS entries as appropriate)")
var interactiveFlag = flag.Bool("interactive", false, "Execute in interactive mode (review proposed update, then prompt user, then create/update/archive Catalog OSS entries as appropriate)")
var resetFlag = flag.Bool("reset", false, "Delete previously existing Catalog OSS entries matching the service name")
var forceFlag = flag.Bool("force", false, "Force rewriting existing Catalog OSS entries even if they are not modified")
var serviceName = flag.String("service", "", "Name of one single service or component to be updated")
var cosEndpoint = flag.String("cosendpoint", "",
	"Endpoints are used hand in hand with your credentials (i.e. keys, CRN, bucket name) to tell your service where to look for this bucket. "+
		"Depending on where your service or applications is located you will want to use one of the below endpoint types. "+
		"For example : s3.us-east.cloud-object-storage.appdomain.cloud")
var cosBucketName = flag.String("cosbucketname", "", "Name of COS bucket. For example : rmc-data")
var cosEmergBucketName = flag.String("cosemergencybucketname", "", "Name of Emergency COS bucket. For example : oss-rmc-emergency")
var cosBucketLink = flag.String("cosbucketlink", "", "A href link of COS bucket. For example : https://cloud.ibm.com/objectstorage/crn%3Av1%3Abluemix%3Apublic%3Acloud-object-storage%3Aglobal%3Aa%2F0bb4d59c58f057ca240dd82f9bc42eb3%3A32d69b06-6a20-4418-88cd-8605d38ff183%3A%3A?bucket=rmc-data&bucketRegion=us-east&endpoint=s3.us-east.cloud-object-storage.appdomain.cloud&paneId=bucket_overview")

var patternString = flag.String("pattern", "", "A regular expression pattern for the names of all services/components to be updated or \"ALL\" to update all known services")
var outputFileName = flag.String("output", AppName+"-<mode>-output-<timestamp>.json", "Path to a file in which to store all the new/updated OSS records (in JSON format), regardless of whether they actually got written to the Catalog itself")
var inputFileSource = flag.String("inputfilesource", "local", "Valid options are: local or cos. When the option is cos, information about the cos must be provided. This option will affect all input files")
var validationsFileName = flag.String("validations", AppName+"-<mode>-validations-<timestamp>.txt",
	"Path to a file in which to store all the validation issues encountered during processing")
var logFileName = flag.String("log", AppName+"-<mode>-log-<timestamp>.log", "Path to a file in which to store a log of all the actions and updates performed by this tool")
var backupFileName = flag.String("backup", AppName+"-<mode>-backup-<timestamp>.json", "Path to a file in which to store a backup of the key non-recreatable information from each OSS record")

// osscatimporter: Always use the same name for reports #246
var reportsFileName = flag.String("reportfile", AppName+"-<mode>-<report>.<extension>", `Path to a file in which to store special report output, if any`)
var slackChannel = flag.String("slack-channel", "", "Send notification to a slack channel. For example : rmc-operations")
var slackResponder = flag.String("slack-responder", "", "First responder in slack channel. For example : dpj@us.ibm.com")
var reportsRequested options.ArrayFlag           // -reports flag
var runActionEnabled options.ArrayFlag           // -run flag
var runActionDisabled options.ArrayFlag          // -norun flag
var snImportFileName options.InputFileFlag       // -snimport flag
var legacyInputFile options.InputFileFlag        // -legacy flag
var supportCenterInputFile options.InputFileFlag // -supportcenter flag
var partsInputFile options.InputFileFlag         // -parts flag
var pricingInputFile options.InputFileFlag       // -pricing flag
var clearinghouseInputFile options.InputFileFlag // -clearinghouse flag
func init() {
	flag.Var(&reportsRequested, "reports", `A comma-separated list of special reports to be generated. Currently available reports: "ownership, catalogga, catalogsummaryxl, services4pnp"`)
	flag.Var(&runActionEnabled, "run",
		fmt.Sprintf("A comma-separated list of optional \"run-actions\" that should be included in this run of the %s tool\nAllowed values: %v", AppName, ossrunactions.ListValidRunActionNames()))
	flag.Var(&runActionDisabled, "norun",
		fmt.Sprintf("A comma-separated list of optional \"run-actions\" that should be excluded in this run of the %s tool\nAllowed values: %v", AppName, ossrunactions.ListValidRunActionNames()))
	flag.Var(&snImportFileName, "snimport", "Path to csv file containing ServiceNow CI import file")
	flag.Var(&legacyInputFile, "legacy", `Path to a report from the legacy python-based CRN validation script`)
	flag.Var(&supportCenterInputFile, "supportcenter", `Path to a list of candidates for the Support Center`)
	flag.Var(&partsInputFile, "parts", `Path to a spreadsheet mapping part numbers to product ID, etc.`)
	flag.Var(&pricingInputFile, "pricing", `Path to a JSON dump of Catalog pricing info, created by osscatpricingloader`)
	flag.Var(&clearinghouseInputFile, "clearinghouse", `Path to a csv file exported from ClearingHouse`)
}

//	var serviceNowFileName = flag.String("servicenow", "", "Path do csv file containing a dump of ServiceNow CIs (may override the data obtained from the SN API)")

// store all actions will be take later
var catalogActions []*preparedCatalogAction

type preparedCatalogAction struct {
	target ossmergemodel.Model
	diffs  *compare.Output
	action stats.Action
}

var outputFile *os.File
var validationsFile *os.File
var logFile *os.File
var backupFile *os.File

// Options is a pointer to the global options.
var Options *options.Data

var runMode options.RunMode
var serviceNameMsg string
var pattern *regexp.Regexp
var fullLogFileName string

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
	if *inputFileSource != "local" && *inputFileSource != "cos" {
		usageError("unsupported option for -inputfilesource. The valid options are: local or cos.", true)
	}
	if *inputFileSource == "cos" && (*cosEndpoint == "" || *cosBucketName == "" || *cosEmergBucketName == "") {
		usageError("-cosendpoint, -cosbucketname, and -cosemergencybucketname are required when -inputfilesource is cos", true)
	}
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
			// test connection only. do not download

			//_, err = cos.Download(*cosEndpoint, *cosBucketName, "run-"+AppName+".txt")
			//if err != nil {
			//	usageError("Skip run "+AppName+", for run-"+AppName+".txt file was not found in cos bucket "+*cosBucketName, true)
			//}
		}
		// test emergency bucket
		isPassEmerg, errEmerg := cos.TestConnection(*cosEndpoint, *cosEmergBucketName)
		if !isPassEmerg || errEmerg != nil {
			detailError := ""
			if errEmerg != nil {
				detailError = errEmerg.Error()
			}
			usageError(fmt.Sprintf("Specified COS service is not available(%s %s). Error is : %s", *cosEndpoint, *cosEmergBucketName, detailError), true)
		} else {
			// try downloading
			_, err = cos.Download(*cosEndpoint, *cosEmergBucketName, "run-"+AppName+".txt")
			if err != nil {
				usageError("Skip run "+AppName+", for run-"+AppName+".txt file was not found in cos bucket "+*cosEmergBucketName, true)
			}
		}
	}
	if debug.IsDebugEnabled(debug.Profiling) {
		debug.Info("CPU profiling enabled")
		defer profile.Start().Stop()
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

	// prepare input files
	prepareInputFiles()

	if *serviceName != "" && *patternString != "" {
		usageError("-service and -pattern flags both specified", true)
	}
	switch {
	case *serviceName != "":
		if !ossrunactions.Services.IsEnabled() {
			panic(fmt.Sprintf(`Cannot operate on a single service when the "Services" run option is not specified (service=%s) `, *serviceName))
		}

		serviceNameMsg = fmt.Sprintf("service \"%s\" only", *serviceName)
		pattern = nil // Just to make sure but should already be zero
		if runMode == options.RunModeRW || runMode == options.RunModeInteractive {
			if debug.IsDebugEnabled(debug.Main) {
				debug.PrintError(`(warning) Operating in -%s mode with a single -service entry (cross-entry relationships cannot be determined; overridden with -debug=1 flag)`, runMode.ShortString())
			} else {
				usageError(fmt.Sprintf(`Cannot operate in -%s mode with a single -service entry (cross-entry relationships cannot be determined; override with -debug=1 flag)`, runMode.ShortString()), false)
			}
		}
	case *patternString != "":
		if strings.ToLower(*patternString) == "all" {
			*patternString = ".*"
			serviceNameMsg = fmt.Sprintf("all services and components found from all sources (with visibility=%s and above)", Options.VisibilityRestrictions)
		} else {
			serviceNameMsg = fmt.Sprintf(`services and components matching the pattern "%s" (with visibility=%s and above)`, *patternString, Options.VisibilityRestrictions)
			if runMode == options.RunModeRW || runMode == options.RunModeInteractive {
				if debug.IsDebugEnabled(debug.Main) {
					debug.PrintError(`(warning) Operating in -%s mode even though the -pattern flag specifies a subset of records (cross-entry relationships cannot be determined; overridden with -debug=1 flag)`, runMode.ShortString())
				} else {
					usageError(fmt.Sprintf(`Cannot operate in -%s mode if the -pattern flag specifies a subset of records (cross-entry relationships cannot be determined; override with -debug=1 flag)`, runMode.ShortString()), false)
				}
			}
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

	// Open the log file.
	// This must happen before we start generating any meaningful output
	fullLogFileName = options.GetOutputFileName(*logFileName, runMode, "", "log", "log")
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
			summarizeCountsOutput := summarizeCounts("        ", false)
			debug.Audit(fmt.Sprintf("Completed run of %s (%s mode): \n%s", AppName, runMode.String(), summarizeCountsOutput))
			utils.PostSlackMessage(AppName, *slackChannel, fmt.Sprintf("Completed run of %s (%s mode)", AppName, runMode.String()), summarizeCountsOutput, utils.InfoType, "")
			debug.Audit(debug.SummarizeErrors("        "))
		}
	}()

	if runMode == options.RunModeRW || runMode == options.RunModeInteractive {
		options.BootstrapLogDNA(AppName+`-`+runMode.ShortString(), true, 0)
	}

	err = ossrunactions.Enable(runActionEnabled)
	if err != nil {
		usageError(fmt.Sprintf("Error in \"-run\" flag: %v\n", err), true)
	}
	err = ossrunactions.Disable(runActionDisabled)
	if err != nil {
		usageError(fmt.Sprintf("Error in \"-norun\" flag: %v\n", err), true)
	}

	// Check all the report names
	reportsList := make([]*ossreportsregistry.Report, 0, len(reportsRequested))
	var reportErrors int
	for _, reportName := range reportsRequested {
		report := ossreportsregistry.LookupReport(reportName)
		if report != nil {
			reportsList = append(reportsList, report)
		} else {
			fmt.Printf("Unknown report name: \"%s\"\n", reportName)
			reportErrors++
		}
	}
	if reportErrors > 0 {
		usageError("", true)
	}

	debug.Info("%s Version: %s\n" /*os.Args[0]*/, AppName, AppVersion)
	if *versionFlag {
		return
	}

	debug.Audit("Operating in %s mode on %s\n", runMode, serviceNameMsg)

	// Check OSS entry ownership
	if options.IsCheckOwnerSpecified() {
		debug.Info(`Using -check-owner=%v (explicitly specified on command line)`, Options.CheckOwner)
	} else {
		Options.CheckOwner = true
		debug.Info(`Forcing -check-owner=%v`, Options.CheckOwner)
	}

	if options.GlobalOptions().TestMode {
		debug.Info("Test Mode enabled (include data from test instance of ServiceNow)")
	}

	for _, ra := range ossrunactions.ListEnabledRunActionNames() {
		debug.Info("Optional run action ENABLED:  %s", ra)
	}
	for _, ra := range ossrunactions.ListDisabledRunActionNames() {
		debug.Info("Optional run action DISABLED: %s", ra)
	}

	// Process the ServiceNow import file
	if snImportFileName != "" {
		err := servicenow.ReadServiceNowImportFile(string(snImportFileName))
		if err != nil {
			panic(err)
		}
	}

	// Process the legacy input file
	if legacyInputFile != "" {
		err := legacy.SetCRNValidationFile(string(legacyInputFile))
		if err != nil {
			panic(err)
		}
	}

	// Process the Support Center candidates input file
	if supportCenterInputFile != "" {
		err := supportcenter.SetCandidatesInputFile(string(supportCenterInputFile))
		if err != nil {
			panic(err)
		}
		reportsList = append(reportsList, ossreportsregistry.LookupReport("SupportCenter"))
	}

	// Process the parts input file
	if partsInputFile != "" {
		err := partsinput.ReadPartsInputFile(string(partsInputFile))
		if err != nil {
			panic(err)
		}
	}

	// Process the pricing input file
	if pricingInputFile != "" {
		_, err := ossmerge.LoadPricingFile(string(pricingInputFile))
		if err != nil {
			panic(err)
		}
	}

	// Process the ClearingHouse input file
	if clearinghouseInputFile != "" {
		err := clearinghouse.ReadCHInputFile(string(clearinghouseInputFile))
		if err != nil {
			panic(err)
		}
	}

	if pattern != nil {
		if *resetFlag {
			// TODO: Handle reset processing for multiple entries
			//			panic("Processing (reset) for more than one service not yet implemented")
			debug.Info("\n\n================= Review/Prepare phase: showing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())
			err := catalog.ListOSSEntries(pattern, catalog.IncludeAll, func(r ossrecord.OSSEntry) {
				var m ossmergemodel.Model
				switch r1 := r.(type) {
				case *ossrecordextended.OSSSegmentExtended:
					seg := &ossmerge.SegmentInfo{}
					seg.OSSSegment = r1.OSSSegment
					seg.PriorOSS = r1.OSSSegment
					m = seg
				case *ossrecordextended.OSSTribeExtended:
					tr := &ossmerge.TribeInfo{}
					tr.OSSTribe = r1.OSSTribe
					tr.PriorOSS = r1.OSSTribe
					m = tr
				case *ossrecordextended.OSSEnvironmentExtended:
					env := &ossmerge.EnvironmentInfo{}
					env.OSSEnvironment = r1.OSSEnvironment
					env.PriorOSS = r1.OSSEnvironment
					m = env
				case *ossrecordextended.OSSServiceExtended:
					si := &ossmerge.ServiceInfo{}
					si.OSSService = r1.OSSService
					si.PriorOSS = r1.OSSService
					m = si
				default:
					panic(fmt.Sprintf("Unexpected entry type in reset mode: %T  %+v", r, r))
				}
				prepareCatalogAction(m, true)
			})
			if err != nil {
				debug.Critical("Error loading all source entries for pattern \"%s\" in reset mode: %v", pattern, err)
			}
		} else {
			err := ossmerge.LoadAllEntries(pattern)
			if err != nil {
				debug.Critical("Error loading all source entries for pattern \"%s\": %v", pattern, err)
			} else {
				allMergedEntries, err := ossmerge.MergeAllEntries(pattern)
				if err != nil {
					debug.Critical("Error while merging all OSS entries for pattern \"%s\": %v", pattern, err)
				} else {
					// Write all entries to the Catalog
					sort.Slice(allMergedEntries, func(i, j int) bool {
						return allMergedEntries[i].GlobalSortKey() < allMergedEntries[j].GlobalSortKey()
					})
					debug.Info("\n\n================= Review/Prepare phase: showing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())
					for _, m := range allMergedEntries {
						prepareCatalogAction(m, false)
					}
				}
			}
		}
	} else {
		if *resetFlag {
			// Include services domain overrides so we can validate overrides and write the full record back:
			ossrec, err := catalog.ReadOSSService(ossrecord.CRNServiceName(*serviceName))
			if err != nil {
				debug.Critical("Error reading OSS Record for \"%s\": %v", *serviceName, err)
			} else {
				// Create an empty ServiceInfo record
				// Note we use a ServiceInfo even though no merge, in order to be able to keep track of a list of many services
				si, _ := ossmerge.LookupService(ossmerge.MakeComparableName(string(*serviceName)), true)
				si.OSSService = *ossrec
				si.PriorOSS = *ossrec
				//				si.OSSMergeControl = ossrec.OSSMergeControl
				debug.Info("\n\n================= Review/Prepare phase: showing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())
				prepareCatalogAction(si, true)
			}
		} else {
			if ossrunactions.Tribes.IsEnabled() && ossrunactions.Doctor.IsEnabled() {
				err = ossmerge.LoadScorecardV1SegmentTribes()
				if err != nil {
					// TODO: force read-only mode if error loading segments and tribes
					//PrintError("Error loading Segments and Tribes from ScorecardV1: %v", err)
					panic(fmt.Sprintf("Error loading Segments and Tribes from ScorecardV1: %v", err))
				}
			} else {
				debug.Info("Skip reloading Segment and Tribe Info from ScorecardV1")
			}
			si, err := ossmerge.LoadOneService(string(*serviceName))
			if err != nil {
				debug.Critical("Error loading all source entries for \"%s\": %v", *serviceName, err)
			} else {
				err := ossmerge.MergeOneService(si)
				if err != nil {
					debug.Critical("Error merging service/component for \"%s\": %v", *serviceName, err)
				} else {
					debug.Info("\n\n================= Review/Prepare phase: showing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())
					prepareCatalogAction(si, false)
				}
			}
		}
	}

	performCatalogActions()

	debug.Audit("Total number of calls to the ClearingHouse API: %d", clearinghouse.GetCountClearingHouseAPICalls())
	if clearinghouseInputFile != "" {
		debug.PlainLogEntry(debug.LevelINFO, "", "List of ClearingHouse entries referenced during this run but not found in the ClearingHouse csv export file: \n%s", clearinghouse.DumpMissingCHSummaryEntries())
	}
	rest.PrintTimings()

	if stats.GetGlobalActualStats().NumEntries() > 0 && len(reportsList) > 0 {
		if debug.CountCriticals() == 0 {
			myPattern := pattern
			if myPattern == nil {
				comparableName := ossmerge.MakeComparableName(*serviceName)
				myPattern = regexp.MustCompile(regexp.QuoteMeta(comparableName))
			}
			for _, report := range reportsList {
				fullReportFileName := options.GetOutputFileName(*reportsFileName, runMode, "", report.Name, report.FileExtension)
				debug.Info("Generating report %s to file %s", report.Name, fullReportFileName)
				reportFile, err := os.Create(filepath.Clean(fullReportFileName))
				if err != nil {
					debug.PrintError(`Cannot create report file "%s": %v`, fullReportFileName, err)
				}
				defer func() {
					reportFile.Close()
					if *cosEndpoint != "" {
						isUploaded, err := cos.Overwrite(*cosEndpoint, *cosBucketName, fullReportFileName, cos.ContentType(report.FileExtension))
						if err != nil || !isUploaded {
							errorMsg := "Failed to upload file(" + fullReportFileName + ") to COS " + *cosBucketName
							fmt.Println(errorMsg, err)
							errBody := ""
							if os.Getenv("RUNNING_IN_KUBE") == "yes" {
								errBody += "Pod name may be " + os.Getenv("HOSTNAME") + ".\n"
							}
							if err != nil {
								errBody += err.Error()
							}
							utils.PostSlackMessage(AppName, *slackChannel, errorMsg, errBody, utils.ErrorType, *slackResponder)
						} else {
							messageBody := ""
							if *cosBucketLink != "" {
								messageBody = "<" + *cosBucketLink + "&prefix=" + fullReportFileName + "|Open it in the browser.>"
							}
							utils.PostSlackMessage(AppName, *slackChannel, fullReportFileName+" was uploaded", messageBody, utils.InfoType, "")
						}
					}
				}()
				err = report.ReportFunc(reportFile, Options.LogTimeStamp, myPattern)
				if err != nil {
					errorMsg := fmt.Sprintf(`Error while processing report name: "%s": %v`, report.Name, err)
					debug.PrintError(errorMsg)
					reportFile.WriteString(errorMsg)
					reportFile.WriteString("\n")
				}
			}
		} else {
			debug.Warning("Skipping the generation of reports because of CRITICAL issues in this run")
		}
	}

	// Generate an error message for any unknown Catalog tags that were encountered while merging and reporting on all the entries
	catalog.ListUnknownTags()
}
func prepareInputFiles() {
	if *inputFileSource == "cos" {
		if partsInputFile != "" {
			content, err := cos.Download(*cosEndpoint, *cosBucketName, string(partsInputFile))
			if err != nil {
				usageError(fmt.Sprintf("Unable download parts file from cos %s. Error is : %s", *cosBucketName, err.Error()), true)
			}
			if len(content) == 0 {
				usageError("Unable download parts file from cos. content length is 0", true)
			}
			err = ioutil.WriteFile(string(partsInputFile), content, 0600)
			if err != nil {
				usageError(fmt.Sprintf("Unable save parts file. Error is : %s", err.Error()), true)
			}
		}
		if clearinghouseInputFile != "" {
			content, err := cos.Download(*cosEndpoint, *cosBucketName, string(clearinghouseInputFile))
			if err != nil {
				usageError(fmt.Sprintf("Unable download clearinghouse file from cos %s. Error is : %s", *cosBucketName, err.Error()), true)
			}
			if len(content) == 0 {
				usageError("Unable download clearinghouse file from cos. content length is 0", true)
			}
			err = ioutil.WriteFile(string(clearinghouseInputFile), content, 0600)
			if err != nil {
				usageError(fmt.Sprintf("Unable save clearinghouse file. Error is : %s", err.Error()), true)
			}
		}
	}
}

func runOneReport(reportName, extension string, reportFunc func(w io.Writer) error) {
	fname := options.GetOutputFileName(*reportsFileName, runMode, "", reportName, extension)
	debug.Info("Generating report %s to file %s", reportName, fname)
	reportFile, err := os.Create(filepath.Clean(fname))
	if err != nil {
		debug.PrintError(`Cannot create report file "%s": %v`, fname, err)
	}
	defer func() {
		reportFile.Close()
		if *cosEndpoint != "" {
			isUploaded, err := cos.Overwrite(*cosEndpoint, *cosBucketName, fname, cos.ContentType(extension))
			if err != nil || !isUploaded {
				errorMsg := "Failed to upload file(" + fname + ") to COS " + *cosBucketName
				fmt.Println(errorMsg, err)
				errBody := ""
				if os.Getenv("RUNNING_IN_KUBE") == "yes" {
					errBody += "Pod name may be " + os.Getenv("HOSTNAME") + ".\n"
				}
				if err != nil {
					errBody += err.Error()
				}
				utils.PostSlackMessage(AppName, *slackChannel, errorMsg, errBody, utils.ErrorType, *slackResponder)
			} else {
				messageBody := ""
				if *cosBucketLink != "" {
					messageBody = "<" + *cosBucketLink + "&prefix=" + fname + "|Open it in the browser.>"
				}
				utils.PostSlackMessage(AppName, *slackChannel, fname+" was uploaded", messageBody, utils.InfoType, "")
			}
		}
	}()
	err = reportFunc(reportFile)
	if err != nil {
		errorMsg := fmt.Sprintf(`Error while processing report name: "%s": %v`, reportName, err)
		debug.PrintError(errorMsg)
		reportFile.WriteString(errorMsg)
		reportFile.WriteString("\n")
	}
}

func prepareCatalogAction(model ossmergemodel.Model, reset bool) {
	out := model.Diffs()
	numDiffs := out.NumDiffs()
	// Check if the only difference is the SchemaVersion
	if numDiffs == 1 {
		diffs := out.GetDiffs()
		if diffs[0].LName == "before.SchemaVersion" {
			numDiffs = 0
		}
	}

	var action stats.Action
	if reset {
		if model.HasPriorOSS() {
			action = stats.ActionDelete
		} else {
			action = stats.ActionIgnore(fmt.Sprintf(`entry should be reset did not exist previously: "%s"`, model.String()))
		}
	} else if model.OSSEntry().GetOSSTags().Contains(osstags.OSSDelete) {
		if model.HasPriorOSS() {
			action = stats.ActionDelete
		} else {
			action = stats.ActionIgnore(fmt.Sprintf(`entry has "%s" tag and did not exist previously: "%s"`, osstags.OSSDelete, model.String()))
		}
	} else if !model.IsValid() {
		action = stats.ActionIgnore(fmt.Sprintf(`no name from any valid sources: "%s"`, model.String()))
	} else if !model.IsUpdatable() {
		action = stats.ActionCatalogNative
	} else if model.IsDeletable() {
		if model.HasPriorOSS() {
			action = stats.ActionDelete
		} else {
			action = stats.ActionIgnore(fmt.Sprintf(`entry is deletable and did not exist previously: "%s"`, model.String()))
		}
	} else if model.HasPriorOSS() {
		if runMode == options.RunModeRW {
			if *forceFlag {
				action = stats.ActionUpdate
			} else if numDiffs == 0 {
				action = stats.ActionNotModified
			} else {
				action = stats.ActionUpdate
			}
		} else {
			if numDiffs == 0 {
				action = stats.ActionNotModified
			} else {
				action = stats.ActionUpdate
			}
		}
	} else {
		if runMode == options.RunModeRW {
			if *forceFlag {
				action = stats.ActionCreate
			} else if numDiffs == 0 {
				action = stats.ActionNotModified
			} else {
				action = stats.ActionCreate
			}
		} else {
			if numDiffs == 0 {
				action = stats.ActionNotModified
			} else {
				action = stats.ActionCreate
			}
		}
	}

	// Write an entry in the logFile and stdout
	// We do this *after* the update, so that any errors will show-up before the update message
	var logString = "*** NO LOG ***"
	if logFile != nil || Options.Verbose {
		logBuffer := new(strings.Builder)
		logBuffer.WriteString("\n")
		logBuffer.WriteString("*** ")
		logBuffer.WriteString(action.Message(runMode, *forceFlag))
		logBuffer.WriteString(" --> ")
		if !action.IsActionIgnore() {
			diffs := out.StringWithPrefix("  UPDATED:")
			if len(diffs) > 0 {
				logBuffer.WriteString(model.Header())
				logBuffer.WriteString(fmt.Sprintf("---- BEGIN DIFFS %s\n", out.Summary()))
				logBuffer.WriteString(diffs)
				logBuffer.WriteString("---- END   DIFFS \n")
				logBuffer.WriteString("-- ")
			}
		}
		logBuffer.WriteString(model.Details())
		logBuffer.WriteString("\n")
		logString = logBuffer.String()
		if logFile != nil {
			debug.PlainLogEntry(debug.LevelINFO, model.String(), logString)
		}
	}
	if Options.Verbose {
		fmt.Print(logString)
	} else {
		fmt.Println()
		fmt.Print("*** ", action.Message(runMode, *forceFlag))
		fmt.Print(" --> ")
		fmt.Print(model.Header())
		fmt.Println()
	}
	if !model.IsUpdatable() && numDiffs > 0 {
		debug.PrintError("Found updatable differences in a non-updatable entry: %s", model.String())
	}

	// Update the counts
	stats.GetGlobalPrepareStats().RecordAction(action, runMode, *forceFlag, model.OSSEntry(), model.PriorOSSEntry())
	catalogActions = append(catalogActions, &preparedCatalogAction{target: model, diffs: out, action: action})
}

func performCatalogActions() {
	var err error

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
		summarizeCountsOutput := summarizeCounts("        ", true)
		debug.Info(fmt.Sprintf("\n\nRunning %s in %s mode - review summary for evaluation: \n%s", AppName, runMode.String(), summarizeCountsOutput))
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
				os.Exit(-1)
			case "readonly":
				runMode = options.RunModeRO
				break loop
			}
		}
	}

	debug.Info("\n\n================= Output/Commit phase: finalizing all proposed OSS record updates (%s mode) ===================\n\n", runMode.String())

	// Open the output file
	fullOutputFileName := options.GetOutputFileName(*outputFileName, runMode, "", "output", "json")
	outputFile, err = os.Create(filepath.Clean(fullOutputFileName))
	if err != nil {
		panic(debug.WrapError(err, "Cannot create the output file"))
	}
	io.WriteString(outputFile, fmt.Sprintf("[\n"))
	io.WriteString(outputFile, fmt.Sprintf("  {}")) // Dummy first record, so that we can use a comma to start each new record
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
		uploadFileToCOS(fullOutputFileName, cos.CONTENT_TYPE_JSON)
	}()

	// Open the validations file
	fullValidationsFileName := options.GetOutputFileName(*validationsFileName, runMode, "", "validations", "txt")
	validationsFile, err = os.Create(filepath.Clean(fullValidationsFileName))
	if err != nil {
		panic(debug.WrapError(err, "Cannot create the validations file"))
	}
	defer func() {
		validationsFile.Close()
		uploadFileToCOS(fullValidationsFileName, cos.CONTENT_TYPE_TEXT)
	}()

	if pattern != nil {
		// Open the backup file
		fullBackupFileName := options.GetOutputFileName(*backupFileName, runMode, "", "backup", "json")
		backupFile, err = os.Create(filepath.Clean(fullBackupFileName))
		if err != nil {
			panic(debug.WrapError(err, "Cannot create the backup file"))
		}
		io.WriteString(backupFile, fmt.Sprintf("{\n"))
		io.WriteString(backupFile, fmt.Sprintf(`  "pattern": "%s",`, pattern.String()))
		io.WriteString(backupFile, fmt.Sprintf("\n"))
		io.WriteString(backupFile, fmt.Sprintf(`  "oss_key_backups": [\n`))
		io.WriteString(backupFile, fmt.Sprintf("    {}")) // Dummy first record, so that we can use a comma to start each new record
		defer func() {
			io.WriteString(backupFile, fmt.Sprintf("\n  ]\n"))
			io.WriteString(backupFile, fmt.Sprintf("}\n"))
			// #nosec G307
			backupFile.Close()
			/*
				err := backupFile.Close()
				if err != nil {
					panic(debug.WrapError(err, "Cannot close the backup file"))
				}
			*/
			uploadFileToCOS(fullBackupFileName, cos.CONTENT_TYPE_JSON)
		}()
	}

	//	count := len(catalogActions)
	for _ /* i */, catalogAction := range catalogActions {
		//		fmt.Println("do " + strconv.Itoa(i) + "/" + strconv.Itoa(count) + "...")
		performCatalogAction(catalogAction)
	}
}

func performCatalogAction(preparedCatalogAction *preparedCatalogAction) stats.Action {
	model := preparedCatalogAction.target
	action := preparedCatalogAction.action
	switch action {
	case stats.ActionUpdate:
		if runMode == options.RunModeRW {
			err := catalog.UpdateOSSEntry(model.OSSEntry(), catalog.IncludeAll)
			if err != nil {
				debug.PrintError("Error updating OSS Catalog entry for \"%s\": %v", model.String(), err)
				action = stats.ActionError
			}
		}
	case stats.ActionCreate:
		if runMode == options.RunModeRW {
			err := catalog.CreateOSSEntry(model.OSSEntry(), catalog.IncludeAll)
			if err != nil {
				debug.PrintError("Error creating OSS Catalog entry for \"%s\": %v", model.String(), err)
				action = stats.ActionError
			}
		}
	case stats.ActionDelete:
		if runMode == options.RunModeRW {
			err := catalog.DeleteOSSEntry(model.OSSEntry())
			if err != nil {
				debug.PrintError("Error deleting OSS Catalog entry for \"%s\": %v", model.String(), err)
				action = stats.ActionError
			}
		}
	case stats.ActionCatalogNative:
		// ignore
	}

	// Update the counts
	stats.GetGlobalActualStats().RecordAction(action, runMode, *forceFlag, model.OSSEntry(), model.PriorOSSEntry())

	// Write an entry in the logFile and stdout
	// We do this *after* the update, so that any errors will show-up before the update message
	logBuffer := new(strings.Builder)
	logBuffer.WriteString("*** ")
	logBuffer.WriteString(action.Message(runMode, *forceFlag))
	logBuffer.WriteString(" --> ")
	logBuffer.WriteString(model.Header())
	logString := logBuffer.String()
	if logFile != nil {
		debug.PlainLogEntry(debug.LevelINFO, model.String(), logString)
	}
	fmt.Print(logString)

	if /*action == stats.ActionError ||*/ action == stats.ActionDelete || action.IsActionIgnore() {
		return action
	}

	switch m1 := model.(type) {
	case *ossmerge.ServiceInfo:
		// Write an entry in the outputFile
		if outputFile != nil {
			io.WriteString(outputFile, ",\n")
			io.WriteString(outputFile, m1.OSSServiceExtended.JSON())
		}
		// Write an entry in the validationsFile
		if validationsFile != nil {
			io.WriteString(validationsFile, "*** ")
			io.WriteString(validationsFile, m1.OSSService.Header())
			io.WriteString(validationsFile, m1.OSSValidation.Details())
			io.WriteString(validationsFile, "\n\n")
		}
		// Write an entry in the backupFile
		if backupFile != nil {
			k := ossbackup.NewKeyBackup(&m1.OSSServiceExtended)
			io.WriteString(backupFile, ",\n")
			io.WriteString(backupFile, "    ")
			io.WriteString(backupFile, k.String())
		}
	case *ossmerge.SegmentInfo:
		// Write an entry in the outputFile
		if outputFile != nil {
			io.WriteString(outputFile, ",\n")
			io.WriteString(outputFile, m1.OSSSegmentExtended.JSON())
		}
		// Write an entry in the validationsFile
		if validationsFile != nil {
			if m1.OSSValidation != nil {
				io.WriteString(validationsFile, "*** ")
				io.WriteString(validationsFile, m1.OSSSegment.Header())
				io.WriteString(validationsFile, m1.OSSValidation.Details())
				io.WriteString(validationsFile, "\n\n")
			}
		}
		// TODO: backupFile for OSSSegment
	case *ossmerge.TribeInfo:
		// Write an entry in the outputFile
		if outputFile != nil {
			io.WriteString(outputFile, ",\n")
			io.WriteString(outputFile, m1.OSSTribeExtended.JSON())
		}
		// Write an entry in the validationsFile
		if validationsFile != nil {
			if m1.OSSValidation != nil {
				io.WriteString(validationsFile, "*** ")
				io.WriteString(validationsFile, m1.OSSTribe.Header())
				io.WriteString(validationsFile, m1.OSSValidation.Details())
				io.WriteString(validationsFile, "\n\n")
			}
		}
		// TODO: backupFile for OSSTribe
	case *ossmerge.EnvironmentInfo:
		// Write an entry in the outputFile
		if outputFile != nil {
			io.WriteString(outputFile, ",\n")
			io.WriteString(outputFile, m1.OSSEnvironmentExtended.JSON())
		}
		// Write an entry in the validationsFile
		if validationsFile != nil {
			io.WriteString(validationsFile, "*** ")
			io.WriteString(validationsFile, m1.OSSEnvironment.Header())
			io.WriteString(validationsFile, m1.OSSValidation.Details())
			io.WriteString(validationsFile, "\n\n")
		}
		// TODO: backupFile for OSSEnvironment
	default:
		panic(fmt.Sprintf("Unexpected type: %#v", model))
	}

	return action
}

func summarizeCounts(prefix string, prepare bool) string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("%s%-52s: %v\n", prefix, "Optional run actions ENABLED", ossrunactions.ListEnabledRunActionNames()))
	buf.WriteString(fmt.Sprintf("%s%-52s: %v\n", prefix, "Optional run actions DISABLED", ossrunactions.ListDisabledRunActionNames()))
	buf.WriteString(fmt.Sprintf("%s%-52s: %d\n", prefix, "Number of calls to the ClearingHouse API", clearinghouse.GetCountClearingHouseAPICalls()))
	if prepare {
		buf.WriteString(stats.GetGlobalPrepareStats().Report(prefix))
	} else {
		buf.WriteString(stats.GetGlobalActualStats().Report(prefix))
	}

	return buf.String()
}

func uploadFileToCOS(fullFileName, contentType string) {
	if *cosEndpoint != "" {
		isUploaded, err := cos.Upload(*cosEndpoint, *cosBucketName, fullFileName, contentType)
		if err != nil || !isUploaded {
			errTitle := "Failed to upload file(" + fullFileName + ") to COS " + *cosBucketName
			errBody := ""
			if os.Getenv("RUNNING_IN_KUBE") == "yes" {
				errBody += "Pod name may be " + os.Getenv("HOSTNAME") + ".\n"
			}
			if err != nil {
				errBody += err.Error()
			}
			fmt.Println(errTitle, errBody)
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
