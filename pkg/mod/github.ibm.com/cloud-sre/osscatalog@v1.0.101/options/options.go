package options

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	cfenv "github.com/cloudfoundry-community/go-cfenv"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// Global options for the oss-catalog library and tools

// Data is a structure containing all the options
type Data struct {
	Config                    string
	Verbose                   bool
	Debug                     int
	KeyFile                   string
	VcapFile                  string
	Lenient                   bool
	TestMode                  bool
	LogTimeStamp              string
	CFEnv                     *cfenv.App
	CFServices                cfenv.Services
	VisibilityRestrictions    catalogapi.VisibilityRestrictions
	IncludePricing            bool
	RefreshPricing            bool
	rawVisibilityRestrictions string
	productionWriteEnabled    string
	NoLogDNA                  bool
	CheckOwner                bool
}

const productionWriteKey = "Riker-Omega-3"

var options *Data
var defaultOptions *Data

// GlobalOptions returns the options Data structure containing all the options in effect
func GlobalOptions() *Data {
	// Return options if set:
	if options != nil {
		return options
	}
	// options is not set, return default options:
	if defaultOptions == nil {
		defaultOptions = createDefaultGlobalOptions()
	}
	return defaultOptions
}

const checkOwnerFlag = "check-owner"

// Returns Data structure with the default options set
func createDefaultGlobalOptions() *Data {
	defaultGlobalOptions := Data{}
	defaultGlobalOptions.Config = ""
	defaultGlobalOptions.Verbose = false
	defaultGlobalOptions.Debug = 0
	defaultGlobalOptions.KeyFile = ""
	defaultGlobalOptions.VcapFile = ""
	defaultGlobalOptions.Lenient = false
	defaultGlobalOptions.TestMode = false
	defaultGlobalOptions.IncludePricing = false
	defaultGlobalOptions.RefreshPricing = false
	defaultGlobalOptions.rawVisibilityRestrictions = string(catalogapi.VisibilityPrivate)
	defaultGlobalOptions.productionWriteEnabled = ""
	defaultGlobalOptions.NoLogDNA = false
	defaultGlobalOptions.CheckOwner = false
	return &defaultGlobalOptions
}

// LoadGlobalOptions loads and validates all the global options,
// and returns the OptionsData structure.
// If "args" is non-empty string, the options are loaded from it; otherwise they are loaded from the command line
// If "reset" is true, allow this function to be called more than once (used for testing)
func LoadGlobalOptions(args string, reset bool) *Data {
	var err error
	if options == nil {
		// First time through. Perform non-reentrant initializations
		// flag.Parsed() is automatically called during Go test initialization
		//	if flag.Parsed() {
		//		panic("flag.Parse() called before call to options.LoadGlobalOptions()")
		//	}

		options = &Data{}
		flag.StringVar(&options.Config, "config", "", "Path to file containing command-line parameters (.yaml)")
		flag.BoolVar(&options.Verbose, "verbose", false, "Verbose output")
		flag.IntVar(&options.Debug, "debug", 0, "Debugging flags")
		flag.StringVar(&options.KeyFile, "keyfile", "", "Path to file containing authentication keys")
		flag.StringVar(&options.VcapFile, "vcap", "", "Path to file VCAP_SERVICES info (for testing)")
		flag.BoolVar(&options.Lenient, "lenient", false, "Load OSS records even if they are incomplete -- used to upgrade to new record formats")
		flag.BoolVar(&options.TestMode, "test", false, "Run in test mode (include input from test instance of ServiceNow, etc.")
		flag.BoolVar(&options.IncludePricing, "include-pricing", false, "Load pricing info (part numbers) from the Main records in Catalog -- only if not already present in prior OSS records")
		flag.BoolVar(&options.RefreshPricing, "refresh-pricing", false, "Refresh pricing info (part numbers) from the Main records in Catalog -- even if already present in prior OSS records")
		flag.StringVar(&options.rawVisibilityRestrictions, "visibility", string(catalogapi.VisibilityPrivate),
			fmt.Sprintf("Limit the scan to Catalog records with public or IBM visibility (%s, %s or %s) or above", catalogapi.VisibilityPublic, catalogapi.VisibilityIBMOnly, catalogapi.VisibilityPrivate))
		flag.BoolVar(&options.NoLogDNA, "no-logdna", false, "Do not write to LogDNA")
		flag.BoolVar(&options.CheckOwner, checkOwnerFlag, false, "Reject OSS entries not owned by osscat@us.ibm.com")
		//	flag.StringVar(&options.LogTimeStamp, "logtimestamp", "xx-xx-xx", "Timestamp to include in file names for output files") // This is automatically computed
	} else if reset {
		// TODO: Should not have to copy all the default values again
		*options = *createDefaultGlobalOptions()
	} else {
		panic("options.LoadGlobalOptions() called more than once")
	}

	if args == "" {
		// TODO: Can we actually call flag.Parse() more than once?
		flag.Parse() // Parse from the command line
	} else {
		err := flag.CommandLine.Parse(strings.Split(args, " "))
		if err != nil {
			debug.PrintError("%v", err)
			//			flag.PrintDefaults()
			os.Exit(2)
		}
	}

	if options.Config != "" {
		err := ReadConfigFile(options.Config)
		if err != nil {
			debug.PrintError("%v", err)
			//			flag.PrintDefaults()
			os.Exit(2)
		}
	}

	// Compute the LogTimeStamp
	timeLocation := func() *time.Location {
		loc, _ := time.LoadLocation("UTC")
		return loc
	}()
	options.LogTimeStamp = time.Now().In(timeLocation).Format("2006-01-02T1504Z")

	// Handle debug flags
	if debugEnv := os.Getenv("OSSCAT_DEBUG"); debugEnv != "" {
		v, err := strconv.ParseInt(debugEnv, 0, 32)
		if err == nil {
			options.Debug += int(v)
		} else {
			debug.PrintError(`Cannot parse OSSCAT_DEBUG="%s" environment variable: %v`, debugEnv, err)
		}
	}
	if options.Debug != 0 {
		debug.Info("Setting debug flags to 0x%x\n", options.Debug)
		debug.SetDebugFlags(options.Debug)
	}

	// Load the CF environment info
	if options.VcapFile != "" {
		debug.Info("Getting VCAP information from file %s", options.VcapFile)
		file, err := os.Open(options.VcapFile)
		if err != nil {
			panic(fmt.Sprintf("Cannot open vcap file %s: %v", options.VcapFile, err))
		}
		defer file.Close() // #nosec G307
		rawData, err := ioutil.ReadAll(file)
		if err != nil {
			panic(fmt.Sprintf("Error reading vcap file %s: %v", options.VcapFile, err))
		}
		err = json.Unmarshal(rawData, &options.CFServices)
		if err != nil {
			panic(fmt.Sprintf("Error parsing vcap file %s: %v", options.VcapFile, err))
		}
		options.CFEnv = nil
	} else if cfenv.IsRunningOnCF() {
		options.CFEnv, err = cfenv.Current()
		if err != nil {
			panic(fmt.Sprintf("Error loading the CF environment: %v", err))
		}
		options.CFServices = options.CFEnv.Services
	}

	// Load the KeyFile or look for a VCAP_SERVICES environment variable that contains key information, or look for keys in environment variables
	if options.KeyFile != "" {
		if strings.ToLower(options.KeyFile) == "default" {
			err := rest.LoadDefaultKeyFile()
			if err != nil {
				panic(fmt.Sprintf("Error reading DEFAULT keyfile: %v", err))
			}
		} else if strings.ToLower(options.KeyFile) == "environment" || strings.ToLower(options.KeyFile) == "env" {
			err := rest.LoadEnvironmentSecrets()
			if err != nil {
				panic(fmt.Sprintf("Error reading secrets from the process environment: %v", err))
			}
		} else if options.KeyFile == "<none>" { // Special case for testing
			// ignore
		} else {
			err := rest.LoadKeyFile(options.KeyFile)
			if err != nil {
				panic(fmt.Sprintf("Error reading keyfile: %v", err))
			}
		}
	} else if options.CFServices != nil {
		debug.Info("Getting authentication keys from VCAP_SERVICES")
		err = rest.LoadVCAPServices(options.CFServices)
		if err != nil {
			panic(fmt.Sprintf("Error loading keys from CF VCAP_SERVICES: %v", err))
		}
	} else {
		fmt.Println("-keyfile not specified")
		fmt.Println()
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	// Check the -visibility flag
	options.VisibilityRestrictions, err = catalogapi.ParseVisibilityRestrictions(options.rawVisibilityRestrictions)
	if err != nil {
		fmt.Println(err)
		fmt.Println()
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	return options
}

var modeRegexp = regexp.MustCompile("<mode>")
var reportNameRegexp = regexp.MustCompile("<report>")
var extensionRegexp = regexp.MustCompile("<extension>")
var timeStampRegexp = regexp.MustCompile("<timestamp>")

// GetOutputFileName constructs a file name, substituting a report name and timestamp in the fileNameTemplate
func GetOutputFileName(fileNameTemplate string, runMode RunMode, tag string, reportName, extension string) string {
	str := reportNameRegexp.ReplaceAllLiteralString(fileNameTemplate, reportName)
	str = extensionRegexp.ReplaceAllLiteralString(str, extension)
	str = timeStampRegexp.ReplaceAllLiteralString(str, options.LogTimeStamp)
	str = modeRegexp.ReplaceAllLiteralString(str, runMode.ShortString()+tag)
	return str
}

// SetProductionWriteEnabled set the global flag to enable writes to the Production Global Catalog instance
func SetProductionWriteEnabled(f bool) {
	if f {
		debug.Info("*** Set ProductionWriteEnabled flag: TRUE")
		GlobalOptions().productionWriteEnabled = productionWriteKey
	} else {
		debug.Info("SetProductionWriteEnabled(%v)", f)
		GlobalOptions().productionWriteEnabled = ""
	}
}

// IsProductionWriteEnabled returns true if the global option to enable writes to the Production Global Catalog instance
// is enabled
func IsProductionWriteEnabled() bool {
	if GlobalOptions().productionWriteEnabled == productionWriteKey {
		return true
	}
	return false
}

// IsCheckOwnerSpecified returns true if the "check-owner" flag was specified explicitly on the command line, as opposed to simply having its default value
func IsCheckOwnerSpecified() bool {
	var found bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == checkOwnerFlag {
			found = true
		}
	})
	return found
}
