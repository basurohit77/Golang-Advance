package debug

// debug.go contains utility functions for debugging and tracking errors

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Constant declarations for debug flags
// TODO: Avoid declaring flags for the osslibs package in this osscatalog package
const (
	Main          = 1 << iota // 0x1 debugging for main phases of the program
	Fine                      // 0x2 fine-grained debugging
	SysID                     // 0x4 debugging for the ServiceNow CI-sysid mapping
	ServiceNow                // 0x8 debugging for the ServiceNow API access
	Catalog                   // 0x10 debugging for the Global Catalog access - including all read accesses during loading phase)
	ScorecardV1               // 0x20 debugging for the Doctor ScorecardV1 access
	IAM                       // 0x40 debugging for accesses to IAM
	Reports                   // 0x80 debugging for reports
	Visibility                // 0x100 debugging for Catalog visibility
	MergeControl              // 0x200 debugging for OSSMergeControl
	AppID                     // 0x400 debugging for AppID interactions and authentication
	Merge                     // 0x800 debugging for merge functions
	Options                   // 0x1000 debugging for options package
	Pricing                   // 0x2000 debugging for Pricing info from Catalog
	ClearingHouse             // 0x4000 debugging for ClearingHouse linkage
	LogDNA                    // 0x8000 debugging for LogDNA integration
	Composite                 // 0x10000 debugging for handling of "composite" entries in Catalog
	Monitoring                // 0x20000 debugging for the loading/merging of availability monitoring meta-data
	Environments              // 0x40000 debugging for OSSEnvironment entries
	Doctor                    // 0x80000 debugging for access to Doctor
	CatalogWrite              // 0x100000 debugging for write access to Global Catalog (excluding read access during the loading phase)
	Profiling                 // 0x200000 profiling execution
	RMC                       // 0x400000 debugging for access to RMC API
	Compare                   // 0x800000 debugging for the `compare` package
)

var _ = os.Stdout      // Force loading of os package even if not used because logging is disabled
var _ = ioutil.Discard // Force loading os ioutil package for logging

var mainOutputChannel = &multiWriter{}

var logChannel io.Writer

func init() {
	mainOutputChannel.AddWriter(os.Stdout)
}

var debugLogger = log.New(os.Stdout, "DEBUG ", log.Ltime)
var errorLogger = log.New(mainOutputChannel, "ERROR ", log.Ltime)
var warningLogger = log.New(mainOutputChannel, "WARN ", log.Ltime)
var infoLogger = log.New(mainOutputChannel, "INFO  ", log.Ltime)
var auditLogger = log.New(mainOutputChannel, "AUDIT ", log.Ltime)
var criticalLogger = log.New(mainOutputChannel, "CRITICAL ", log.Ltime)

var debugFlags = 0

// GetRawOutputChannel returns a direct handle to the io,Writer (or MultiWriter) to which all output is being sent
func GetRawOutputChannel() io.Writer {
	return mainOutputChannel
}

// GetDebugFlags returns the current value of the flags that control the debugging output produced by this program
func GetDebugFlags() int {
	return debugFlags
}

// SetDebugFlags sets the flags that control the debugging output produced by this program
// Returns the previous value of the debug flags
func SetDebugFlags(flags int) int {
	if flags&Catalog != 0 {
		flags |= CatalogWrite
	}
	previous := debugFlags
	debugFlags = flags
	return previous
}

// IsDebugEnabled returns true if debugging is currently enabled for a given debug flag
func IsDebugEnabled(flags int) bool {
	if (flags & debugFlags) == flags {
		return true
	}
	return false
}

// Debug logs one line of debug information, controlled by the debug flags currently enabled
func Debug(flags int, format string, parms ...interface{}) {
	if (flags & debugFlags) == flags {
		debugLogger.Printf(format, parms...)
	}
}

// InfoWithOptions logs one line of general output information, associated with one particular OSSEntryID,
// and returns an error rather than silently proceeding (or panicking)
func InfoWithOptions(id string, format string, parms ...interface{}) error {
	msg := fmt.Sprintf(format, parms...)
	infoLogger.Print(msg)
	err := sendToLogDNA(LevelINFO, false, id, msg)
	return err
}

// Info logs one line of general output information
func Info(format string, parms ...interface{}) {
	msg := fmt.Sprintf(format, parms...)
	infoLogger.Print(msg)
	err := sendToLogDNA(LevelINFO, false, "", msg)
	if err != nil {
		errorList = append(errorList, err.Error())
		errorLogger.Println(err)
	}
}

// AuditWithOptions logs one line of audit information, associated with one particular OSSEntryID,
// and returns an error rather than silently proceeding (or panicking)
func AuditWithOptions(id string, format string, parms ...interface{}) error {
	msg := fmt.Sprintf(format, parms...)
	auditLogger.Print(msg)
	err := sendToLogDNA(LevelAUDIT, true, id, msg)
	return err
}

// Audit logs one line of audit information
func Audit(format string, parms ...interface{}) {
	msg := fmt.Sprintf(format, parms...)
	auditLogger.Print(msg)
	err := sendToLogDNA(LevelAUDIT, true, "", msg)
	if err != nil {
		errorList = append(errorList, err.Error())
		errorLogger.Println(err)
		return
	}
}

// PlainLogEntry logs one line of general output information to the main log file,
// bypassing the normal logger and without including a timestamp.
// It also sends the output to LogDNA if appropriate (this time with a timestamp, which is always required in LogDNA)
func PlainLogEntry(level LogLevel, id string, format string, parms ...interface{}) error {
	var flush bool
	if level == LevelAUDIT || level == LevelERROR {
		flush = true
	}
	msg := fmt.Sprintf(format, parms...)
	_, err := io.WriteString(logChannel, msg)
	if err != nil {
		return err
	}
	if level != "" && level != LevelINFO {
		err = sendToLogDNA(level, flush, id, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

var errorList = make([]string, 0)

// PrintError prints an error message at the point where the error occurred then continues program execution
func PrintError(format string, parms ...interface{}) {
	err := fmt.Errorf(format, parms...)
	msg := err.Error()
	if isKnownError(msg) {
		Warning(msg)
	} else {
		errorList = append(errorList, msg)
		errorLogger.Println(err)
		err2 := sendToLogDNA(LevelERROR, true, "", msg)
		if err2 != nil {
			errorList = append(errorList, err2.Error())
			errorLogger.Println(err2)
		}
	}
}

var criticalList = make([]string, 0)

// Critical prints a critical error message at the point where the error occurred then continues program execution
func Critical(format string, parms ...interface{}) {
	err := fmt.Errorf(format, parms...)
	msg := err.Error()
	criticalList = append(criticalList, msg)
	criticalLogger.Println(err)
	err2 := sendToLogDNA(LevelCRITICAL, true, "", msg)
	if err2 != nil {
		criticalList = append(criticalList, err2.Error())
		criticalLogger.Println(err2)
	}
}

var warningList = make([]string, 0)

// Warning prints a warning message at the point where the error occurred then continues program execution
func Warning(format string, parms ...interface{}) {
	err := fmt.Errorf(format, parms...)
	msg := err.Error()
	warningList = append(warningList, msg)
	warningLogger.Println(err)
	err2 := sendToLogDNA(LevelWARN, true, "", msg)
	if err2 != nil {
		errorList = append(errorList, err2.Error())
		errorLogger.Println(err2)
	}
}

// WrapError wraps a raw error into some context, to pass back up the call stack
func WrapError(err error, message string, parms ...interface{}) error {
	str := fmt.Sprintf(message, parms...)
	err2 := fmt.Errorf("%s: %w", str, err)
	// OK to call Debug here but not another Error function, which could cause a loop calling LogDNA
	Debug(Main, "WrapError: %v", err2)
	return err2
}

// SummarizeErrors summarizes all the errors encountered during a run of this program
func SummarizeErrors(prefix string) string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("%sProgram terminated - %d critical  /  %d errors  /  %d warnings:\n", prefix, len(criticalList), len(errorList), len(warningList)))
	for _, e := range warningList {
		buf.WriteString(fmt.Sprintf("%s    WARNING:  %s\n", prefix, e))
	}
	for _, e := range errorList {
		buf.WriteString(fmt.Sprintf("%s    ERROR:    %s\n", prefix, e))
	}
	for _, e := range criticalList {
		buf.WriteString(fmt.Sprintf("%s    CRITICAL: %s\n", prefix, e))
	}
	return buf.String()
}

// CountErrors returns the number of errors generated during this run
func CountErrors() int {
	return len(errorList)
}

// CountWarnings returns the number of errors generated during this run
func CountWarnings() int {
	return len(warningList)
}

// CountCriticals returns the number of critical issues generated during this run
func CountCriticals() int {
	return len(criticalList)
}

// multiWriter is used to implement writing log and error messages to multiple channels
// include os.Stdout and a log file
type multiWriter struct {
	writers []io.Writer
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	// TODO: Deal with issues if multiple writers do not return the same number of bytes written
	// TODO: Deal with atomicity when writing to multiplle writers
	for i, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return 0, WrapError(err, "MultiWriter.Write error on Writer#%d (%v)", i, w)
		}
	}
	return n, nil
}

// AddWriter adds one Writer to the list of Writers to which this multiWriter sends its output
func (mw *multiWriter) AddWriter(w io.Writer) {
	mw.writers = append(mw.writers, w)
}

// SetLogFile specifies the file (Writer) that will be used as the main log for this program
func SetLogFile(w io.Writer) {
	mainOutputChannel.AddWriter(w)
	logChannel = w
}
