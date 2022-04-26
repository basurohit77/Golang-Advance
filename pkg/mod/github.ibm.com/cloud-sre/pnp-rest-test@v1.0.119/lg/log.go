package lg

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	// ShowErrorLogs controls if error logs are sent to the console
	ShowErrorLogs = true
	// ShowWarningLogs controls if warning logs are sent to the console
	ShowWarningLogs = true
	// ShowInfoLogs controls if info logs are sent to the console
	ShowInfoLogs = true
	// TotalErrors counts the total errors issued
	TotalErrors = 0
	// EnableErrorLog allows error logs to be displayed
	EnableErrorLog = true
	// EnableWarningLog allows warning logs to be displayed
	EnableWarningLog = true
	// EnableInfoLog allows info logs to be displayed
	EnableInfoLog = true
)

// DisableLogLevel will enable logs at a level.  Maybe a comma separated string
func DisableLogLevel(levelstr string) {
	levels := strings.Split(levelstr, ",")
	for _, l := range levels {
		switch l {
		case "info":
			EnableInfoLog = false
		case "error":
			EnableErrorLog = false
		case "warning":
			EnableWarningLog = false
		default:
		}
	}
}

// Err is a logging function for error logs
func Err(fct string, errIn error, msg ...interface{}) (wrapped error) {
	TotalErrors++
	if EnableErrorLog {
		return logIt("ERROR", ShowErrorLogs, fct, errIn, true, msg...)
	}
	return errors.New("Errors Disabled")
}

// Warn is a logging function for warning logs
func Warn(fct string, errIn error, msg ...interface{}) (wrapped error) {
	if EnableWarningLog {
		return logIt("WARN", ShowWarningLogs, fct, errIn, true, msg...)
	}
	return errors.New("Warnings Disabled")
}

// Info is a logging function for informational logs
func Info(fct string, msg ...interface{}) {
	if EnableInfoLog {
		logIt("INFO", ShowInfoLogs, fct, nil, false, msg...)
	}
	return
}

// ShowTestResult will display the result of the tests
func ShowTestResult() {
	if TotalErrors > 0 {
		fmt.Printf("ShowTestResult: FAILURE: Total errors displayed %d\n\n", TotalErrors)
		return
	}
	fmt.Println("ShowTestResult SUCCESS!!!!")
}

func logIt(level string, show bool, fct string, errIn error, getErr bool, msg ...interface{}) (wrapped error) {
	logmsg := getLogMsg(level, fct, errIn, msg...)

	if errIn != nil || getErr {
		wrapped = errors.New(logmsg)
	}

	if show {
		log.Println(logmsg)
	}

	return wrapped
}

func getLogMsg(level string, fct string, errIn error, msg ...interface{}) string {
	logmsg := fmt.Sprintf("[%s][%s]", level, fct)

	if len(msg) > 0 {
		if len(msg) > 1 {
			s := fmt.Sprintf(msg[0].(string), msg[1:]...)
			logmsg += fmt.Sprintf(" %s", s)
		} else {
			logmsg += fmt.Sprintf(" %s", msg[0])
		}
	}

	if errIn != nil {
		logmsg += fmt.Sprintf(" (error:%s)", errIn.Error())
	}
	return logmsg
}
