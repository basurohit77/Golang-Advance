package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.ibm.com/cloud-sre/oss-globals/consts"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"

	osscatcrn "github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
)

// PrettyPrintJSON Prints formatted json to log
//
//prefix: Custom message send to the log ;
//j: Any type of struct
//
// Example: PrettyPrintJSON(tlog.Log()+"ImplList in response body ", riList)
func PrettyPrintJSON(prefix string, j interface{}) {
	b, err := json.MarshalIndent(j, "", " ")
	if err == nil {
		log.Printf("%s:\n%s", prefix, b)
	} else {
		log.Printf("\n %s Failed pretty printing json, err=%v\n", prefix, err)
	}
}

// CreateEndpointHandler - creates an endpoint handler object
func CreateEndpointHandler(path string, handlerFunc func(resp http.ResponseWriter, req *http.Request), validMethods map[string]string) *EndpointHandler {
	endpointHandler := new(EndpointHandler)
	endpointHandler.Path = path
	endpointHandler.HandlerFunc = handlerFunc
	endpointHandler.ValidMethods = validMethods
	return endpointHandler
}

// HandleUnixSignals - handle Unix signals
func HandleUnixSignals() {
	// handle unix signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// This goroutine executes a blocking receive for signals.
	// When it gets one it’ll print it out and then notify the program that it can finish.
	// Note: it will intercept Ctrl-C, "kill <pid>", but NOT "kill -9 <pid>"
	go func() {
		sig := <-sigs
		log.Printf("%s SIGNAL Received: '%s'. Stopping service.\n", tlog.Log(), sig)
		time.Sleep(5 * time.Second)
		os.Exit(-3)
	}()
}

// HandlePanics - intercept panics: print error and stacktrace
func HandlePanics(resp http.ResponseWriter) {
	//const FCT = "HandlePanics: "
	if err := recover(); err != nil {
		stacktraceBuffer := make([]byte, 4096)
		count := runtime.Stack(stacktraceBuffer, true)

		log.Printf(tlog.Log()+"PANIC: '%s'.\n", err)
		log.Printf(tlog.Log()+"STACKTRACE (%d bytes): %s\n", count, stacktraceBuffer[:count])

		msg := fmt.Sprintf(tlog.Log()+"PANIC: '%s'.\n\nSTACKTRACE (%d bytes): %s", err, count, stacktraceBuffer[:count])
		log.Println(msg)
		http.Error(resp, msg, http.StatusInternalServerError)
		return
	}
}

// HandlePanicsForServer - intercept panics: print error and stacktrace
func HandlePanicsForServer() {
	if err := recover(); err != nil {
		stacktraceBuffer := make([]byte, 4096)
		count := runtime.Stack(stacktraceBuffer, true)

		log.Printf(tlog.Log()+"PANIC: '%s'. Stopping service.\n", err)
		log.Printf(tlog.Log()+"STACKTRACE (%d bytes): %s\n", count, stacktraceBuffer[:count])

		time.Sleep(5 * time.Second)
		os.Exit(-2)
	}
}

// GetIDFromURL - Returns the last path in the url string that is not equal to the test path
func GetIDFromURL(urlPath string, testPath string) string {
	id := ""
	lastSegment := ""
	segments := strings.Split(urlPath, "/")
	lastSegment = segments[len(segments)-1]
	if lastSegment == "affects" {
		lastSegment = segments[len(segments)-2]
	}
	if lastSegment != "" && lastSegment != testPath {
		id = lastSegment
	}
	return id
}

// GetServiceFromCRN returns the service name from the provided CRN string, if error return nill and error message
func GetServiceFromCRN(crn string) (service string, err error) {
	crnMask, err := osscatcrn.Parse(crn)

	if err != nil {
		return service, err
	}
	return crnMask.ServiceName, nil
}

// RemoveCRNEntriesAfterLocation returns a CRN with only the portion of the CRN up to and including the location.
// For example, for cr n
//   crn:v1:bluemix:public:tip-api-platform:us-south:a/123:1234-5678-9012-3456:png:CustomerReceipts/clientdinner.png
// the following is returned
//   crn:v1:bluemix:public:tip-api-platform:us-south::::
func RemoveCRNEntriesAfterLocation(crn string) (updatedCRN string, err error) {
	crnMask, err := osscatcrn.ParseAll(crn)

	if err != nil {
		return "", errors.New(tlog.Log() + " Invalid CRN: " + crn)
	}
	separator := consts.CrnSeparator
	return crnMask.Scheme + separator + crnMask.Version + separator + crnMask.CName + separator + crnMask.CType + separator +
		crnMask.ServiceName + separator + crnMask.Location + separator + separator + separator + separator, nil
}

// RemoveServiceNameFromCRN - removes the service name portion from provided crn and returns the updated crn
func RemoveServiceNameFromCRN(crn string) (updatedCRN string, err error) {
	crnMask, err := osscatcrn.ParseAll(crn)

	if err != nil {
		return "", errors.New(tlog.Log() + " Invalid CRN: " + crn)
	}
	separator := consts.CrnSeparator
	return crnMask.Scheme + separator + crnMask.Version + separator + crnMask.CName + separator + crnMask.CType + separator +
		separator + crnMask.Location + separator + crnMask.Scope + separator + crnMask.ServiceInstance + separator +
		crnMask.ResourceType + separator + crnMask.Resource, nil
}

// AddServiceNameToCRN - adds the provided service name to the provided crn and returns the update crn
func AddServiceNameToCRN(crn string, serviceName string) (string, error) {
	if serviceName == "" {
		return crn, nil
	}

	crnMask, err := osscatcrn.ParseAll(crn)
	if err != nil {
		return "", errors.New(tlog.Log() + " Invalid CRN: " + crn)
	}
	separator := consts.CrnSeparator
	return crnMask.Scheme + separator + crnMask.Version + separator + crnMask.CName + separator + crnMask.CType + separator +
		serviceName + separator + crnMask.Location + separator + crnMask.Scope + separator + crnMask.ServiceInstance + separator +
		crnMask.ResourceType + separator + crnMask.Resource, nil
}

// NormalizeCRN - returns an updated version of the crn based on standardization
// if CRN contains `crn:v1:softlayer:` gets replaced with `crn:v1:bluemix:`
func NormalizeCRN(crn string) string {
	crn = strings.ToLower(crn)
	if strings.HasPrefix(crn, "crn:v1:softlayer:") {
		crn = strings.Replace(crn, "crn:v1:softlayer:", "crn:v1:bluemix:", 1)
	}
	return crn
}

// IsPublicCRN - returns true if the provided crn is a public CRN, false otherwise
func IsPublicCRN(crn string) bool {
	crnMask, err := osscatcrn.Parse(crn)
	if err != nil {
		return false
	}
	return crnMask.IsIBMPublicCloud()
}

// IsCrnPnpValid - returns true if the provided crn is a public CRN and GAAS
func IsCrnPnpValid(crn string) bool {
	crnMask, err := osscatcrn.Parse(crn)
	if err != nil {
		return false
	}
	iBMPublic := crnMask.IsIBMPublicCloud()
	log.Print(tlog.Log()+"DEBUG: iBMPublic :", iBMPublic)
	isGaas := IsCrnGaasResource(crn, ctxt.Context{})
	log.Print(tlog.Log()+"DEBUG: isGaas :", isGaas)
	isPnpEnabled, err := osscatalog.IsCRNPnPEnabled(crn)
	if err != nil {
		log.Print(tlog.Log(), "Could not determine if crn is pnp enabled: ", crn, ":", err)
		return false
	}

	if (iBMPublic && isPnpEnabled) || isGaas {
		return true
	}
	return false
}

// IsGaasResource - returns true if the provided OSS Service entry type is a GaaS service, false otherwise
func IsGaasResource(ossServiceEntryType ossrecord.EntryType) bool {
	return ossServiceEntryType == ossrecord.GAAS
}

// IsCrnGaasResource Returns true if CRN is GAAS resource
func IsCrnGaasResource(crn string, context ctxt.Context) bool {

	svc, err := GetServiceFromCRN(crn)
	if err != nil {
		log.Print(tlog.Log(), "DEBUG: IsCrnGaasResource: err: ", err)
		return false
	}
	if svc == "" {
		log.Print(tlog.Log(), "DEBUG: IsCrnGaasResource: service name is empty: ")
		return false
	}

	record, err := osscatalog.ServiceNameToOSSRecord(context, svc)
	if err != nil {
		log.Print(tlog.Log(), "IsCrnGaasResource: ServiceNameToOSSRecord err: ", err)
		return false
	}
	log.Print(tlog.Log(), "DEBUG: IsCrnGaasResource: record.GeneralInfo.EntryType: ", record.GeneralInfo.EntryType)
	return record.GeneralInfo.EntryType == ossrecord.GAAS
}

// IsValidCRN - returns true if the provided crn is valid. that is, if the CRN is of size 10
func IsValidCRN(crn string) bool {
	splitCRN := strings.Split(crn, consts.CrnSeparator)
	return len(splitCRN) == consts.CrnLen
}

// IsRequestContainsParameter - returns true and the parameter value if provided HTTP request contains the parameter in its query parameters, false otherwise
func IsRequestContainsParameter(req *http.Request, parameter string) (bool, string) {
	queryParams := req.URL.Query()
	parameterValues := queryParams[parameter]
	if len(parameterValues) == 0 {
		return false, ""
	}
	return true, parameterValues[0]
}

// IsHeaderContainsValue - returns true and the value if provided HTTP request contains the value in its header, false otherwise
func IsHeaderContainsValue(req *http.Request, value string) (bool, string) {
	headerValue := req.Header.Get(value)
	if headerValue == "" {
		return false, ""
	}
	return true, headerValue
}

// ServerReply replies with a standard way so can be used generically
// Moved from PNP Status
func ServerReply(FCT string, res http.ResponseWriter, reply interface{}) {

	res = defaultHeaders(res)

	log.Println(tlog.Log())

	if err := json.NewEncoder(res).Encode(reply); err != nil {
		log.Println(FCT + err.Error())
	}
	log.Println(tlog.Log(), res)
}

// ServerReplyBytes same as ServerReply but does not need to encode json
func ServerReplyBytes(res http.ResponseWriter, reply []byte) {

	fmt.Fprintf(res, string(reply))
}

// ServerReplyBytesWithStatus same as ServerReply but does not need to encode json and returns specific status code
func ServerReplyBytesWithStatus(res http.ResponseWriter, reply []byte, status int) {

	res.WriteHeader(status)

	fmt.Fprintf(res, string(reply))
}

// Moved from PNP Status
// ErrMsg defines the structure sent in the response if an error occurs
type errMsg struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// ErrHandler replies back with a JSON formated error message
func ErrHandler(res http.ResponseWriter, errString string, status int) {
	var e errMsg
	e.Status = status
	e.Message = errString

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	res.Header().Set("X-ZAP", "OK")
	res.WriteHeader(e.Status)

	if err := json.NewEncoder(res).Encode(e); err != nil {
		log.Println(err)
	}
}

// AttributesToRedact lists all attributes to remove from the data param in redactAttributes()
// it is used for removing sensitive attributes from being exposed in the logs
var AttributesToRedact = []string{"long_description", "new_work_notes", "description", "u_severity", "u_purpose_goal", "backout_plan"}

// RedactAttributes removes sensitive attributes from data
func RedactAttributes(data interface{}) (d []byte) {
	var (
		fct           = "RedactAttributes: "
		dataMap       = make(map[string]interface{})
		isByte, isMap = false, false
		byteData      []byte
		err           error
	)

	// support []byte, map[string]interface{}, and struct types
	switch t := reflect.TypeOf(data).String(); t {
	case "[]uint8": // type []byte
		isByte = true

		// Unmarshal byte array into map
		err = json.Unmarshal(data.([]byte), &dataMap)
		if err != nil {
			log.Println(fct + err.Error())
			return data.([]byte)
		}

	case "map[string]interface {}": // type map[string]interface{}
		isMap = true
		dataMap = data.(map[string]interface{})

	default: // any other type, specifically JSON struct
		byteData, err = json.Marshal(data)
		if err != nil {
			log.Println(fct + err.Error())
			log.Printf("%s%+v\n", fct, data)
			return nil
		}

		// Unmarshal byte array into map
		err = json.Unmarshal(byteData, &dataMap)
		if err != nil {
			log.Println(err)
			return byteData
		}
	}

	var attributesRedacted []string
	for _, attr := range AttributesToRedact {
		if dataMap[attr] != nil {
			attributesRedacted = append(attributesRedacted, attr)
			dataMap[attr] = "redacted"
		}
	}
	log.Println(fct+"redacted attributes:", attributesRedacted)

	d, err = json.Marshal(dataMap)
	if err != nil {
		if isByte {
			return data.([]byte)
		} else if isMap {
			return []byte(fmt.Sprintf("%v", data))
		}
		return byteData
	}

	return d
}
