// osscatviewer: a simple UI / viewer for OSS records in Global Catalog
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/debug"

	"github.ibm.com/cloud-sre/osscatalog/options"
)

// AppVersion is the current version number for the osscatviewer program
const AppVersion = "0.5"

// CookieName is the name of the cookie used to authenticate users who are authorized to update OSSMergeControl records
const CookieName = "osscatviewer-updater"
const dummyAuthToken = "dummy-auth-token" // #nosec G101

var versionFlag = flag.Bool("version", false, "Print the version number of this program")

// Options is a pointer to the global options.
var Options *options.Data

var timeFormat = "Mon Jan 2 15:04:05 MST 2006"

var allowHTTP = flag.Bool("http", false, "Allow requests using http protocol (not https)")
var dummyAuth = flag.Bool("dummy-auth", false, "Simulate an authenticated user")
var authFileName = flag.String("users", "", "Path to file containing a list of authorized users")
var noDiffs = flag.Bool("no-diffs", false, "Do not automatically compute diffs between Staging and Production Catalog")

func main() {
	var err error

	Options = options.LoadGlobalOptions("", false)

	err = options.BootstrapLogDNA("osscatviewer", true, 0)
	if err != nil {
		log.Fatalf("Could not initialize LogDNA: %v", err)
	}

	debug.Info("%s Version: %s  (LogDNA enabled=%v)\n" /*os.Args[0]*/, "osscatviewer", AppVersion, !Options.NoLogDNA)
	if *versionFlag {
		return
	}

	// Check OSS entry ownership
	if options.IsCheckOwnerSpecified() {
		debug.Info(`Using -check-owner=%v (explicitly specified on command line)`, Options.CheckOwner)
	} else {
		Options.CheckOwner = false
		debug.Info(`Forcing -check-owner=%v`, Options.CheckOwner)
	}

	if !*dummyAuth {
		err := setupAuth(*authFileName)
		if err != nil {
			log.Fatalf("Could not initialize the authentication system: %v", err)
		}
	} else {
		debug.PrintError("*** For debugging only: using dummy authentication")
	}

	homeTemplate, err = template.ParseFiles("views/home.html")
	if err != nil {
		log.Fatalf("Could not create the home.html template: %v", err)
	}
	viewFuncs := make(map[string]interface{})
	viewFuncs["getUniversalLink"] = getUniversalLink
	viewTemplate, err = template.New("view.html").Funcs(viewFuncs).ParseFiles("views/view.html")
	if err != nil {
		log.Fatalf("Could not create the view.html template: %v", err)
	}
	tribeTemplate, err = template.ParseFiles("views/tribe.html")
	if err != nil {
		log.Fatalf("Could not create the tribe.html template: %v", err)
	}
	segmentTemplate, err = template.ParseFiles("views/segment.html")
	if err != nil {
		log.Fatalf("Could not create the segment.html template: %v", err)
	}
	segmentsTemplate, err = template.ParseFiles("views/segments.html")
	if err != nil {
		log.Fatalf("Could not create the segments.html template: %v", err)
	}
	environmentTemplate, err = template.ParseFiles("views/environment.html")
	if err != nil {
		log.Fatalf("Could not create the environment.html template: %v", err)
	}
	environmentsTemplate, err = template.ParseFiles("views/environments.html")
	if err != nil {
		log.Fatalf("Could not create the environments.html template: %v", err)
	}
	diffTemplate, err = template.ParseFiles("views/diff.html")
	if err != nil {
		log.Fatalf("Could not create the diff.html template: %v", err)
	}

	refreshFunc := func() {
		err := allRecords.refresh()
		if err != nil {
			log.Fatalf("Error during initial data load from Catalog: %v", err)
		}
	}
	go refreshFunc()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/osscat-styles.css", stylesHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/update/", updateHandler)
	http.HandleFunc("/tribe/", tribeHandler)
	http.HandleFunc("/segment/", segmentHandler)
	http.HandleFunc("/segments/", segmentsHandler)
	http.HandleFunc("/environment/", environmentHandler)
	http.HandleFunc("/environments/", environmentsHandler)
	http.HandleFunc("/diff/", diffHandler)
	http.HandleFunc("/refresh", refreshHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/afterauth", afterAuthHandler)
	port := getPort()
	debug.Info("Listening for requests on port %v   (allowHTTP=%v)", port, *allowHTTP)
	err = http.ListenAndServe(port, nil)
	log.Fatal(err)
}

func getPort() string {
	var port string
	if port = os.Getenv("PORT"); port == "" {
		return ":8080"
	}
	return fmt.Sprintf(":%v", port)
}

func readOneFormField(values url.Values, name string, dest interface{}, errorBuffer *strings.Builder) {
	if inputs, ok := values[name]; ok {
		if len(inputs) != 1 {
			errorBuffer.WriteString(fmt.Sprintf("Expected single value for form field \"%s\" but got %d (%v)\n", name, len(inputs), inputs))
		}
		input := strings.TrimSpace(inputs[0])
		switch d := dest.(type) {
		case *string:
			if len(input) == 0 {
				*d = ""
				return
			}
			switch input[0] {
			case '{', '[', '"', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				break // fall through to json.Unmarshall
			case '*':
				errorBuffer.WriteString(fmt.Sprintf("Found error message in form field \"%s\" (%v)\n", name, input))
				*d = input
				return
			default:
				*d = input
				return
			}
		}
		err := json.Unmarshal([]byte(input), dest)
		if err != nil {
			errorBuffer.WriteString(fmt.Sprintf("Error parsing value for form field \"%s\" (%v): %v\n", name, input, err))
		}
	} else {
		errorBuffer.WriteString(fmt.Sprintf("Missing form field \"%s\"\n", name))
	}
}

func stylesHandler(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := checkRequest(w, r, false); !ok {
		return
	}
	//	w.Header().Set("Content-Type", "text/css")
	http.ServeFile(w, r, "views/osscat-styles.css")
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := checkRequest(w, r, false); !ok {
		return
	}
	http.ServeFile(w, r, "views/favicon.ico")
}

func errorPage(w http.ResponseWriter, statusCode int, msg string, parms ...interface{}) {
	out := fmt.Sprintf("HTTP error %d: %s: %s", statusCode, http.StatusText(statusCode), fmt.Sprintf(msg, parms...))
	debug.PrintError(out)
	w.WriteHeader(statusCode)
	fmt.Fprint(w, out)
}

func checkRequest(w http.ResponseWriter, r *http.Request, privileged bool) (loggedInUser string, updateEnabled bool, ok bool) {
	var tls string
	if r.TLS != nil {
		tls = "yes"
	} else {
		tls = "no"
	}
	var xForwardedProto string
	if hdrs := r.Header["X-Forwarded-Proto"]; hdrs != nil && len(hdrs) > 0 {
		xForwardedProto = hdrs[0]
	}

	loggedInUser, updateEnabled, loginMessage := checkToken(w, r)

	if !*allowHTTP && xForwardedProto != "https" {
		redirectURL := fmt.Sprintf(`https://%s%s`, r.Host, r.URL.Path)
		debug.Info(`Redirecting request - %s "%s" -> "%s"   from="%s"   X-Forwarded-Proto=%s   TLS=%s   updatedEnabled=%v   user=%v`, r.Method, r.URL.String(), redirectURL, r.RemoteAddr, xForwardedProto, tls, updateEnabled, loggedInUser)
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
		return "", false, false
	}
	debug.Info(`Processing request - %s "%s"   from="%s"   X-Forwarded-Proto=%s   TLS=%s   updatedEnabled=%v   user=%v`, r.Method, r.URL.String(), r.RemoteAddr, xForwardedProto, tls, updateEnabled, loggedInUser)

	if privileged && loginMessage != "" {
		debug.Info(loginMessage)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return "", false, false
	}
	if privileged && !updateEnabled {
		errorPage(w, http.StatusUnauthorized, "Must be authorized to edit and update OSS entries")
		return "", false, false
	}

	return loggedInUser, updateEnabled, true
}

// getUniversalLink generates a link (URL) to whatever its target is referencing
// (currently supports OSS entries designated by CRNServiceName or CH entries designated by label)
func getUniversalLink(target string) string {
	if crnServiceNamePattern.MatchString(target) {
		return fmt.Sprintf("/view/%s", target)
	}
	if link := clearinghouse.GetCHEntryUIFromLabel(target); link != "" {
		return link
	}
	return ""
}

var crnServiceNamePattern = regexp.MustCompile(`^[a-z0-9-]*$`)
