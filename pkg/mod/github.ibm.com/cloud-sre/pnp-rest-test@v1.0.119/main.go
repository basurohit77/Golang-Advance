package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/oss-globals/chkdbcon"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-rest-test/catalog"
	"github.ibm.com/cloud-sre/pnp-rest-test/config"
	"github.ibm.com/cloud-sre/pnp-rest-test/forms"
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
	"github.ibm.com/cloud-sre/pnp-rest-test/status"
	"github.ibm.com/cloud-sre/pnp-rest-test/testDefs"
	resourceAdapterTest "github.ibm.com/cloud-sre/pnp-rest-test/testDefs/testDefsResourceAdapter"
)

var (
	hooksHandler                   = testDefs.SimpleRun
	runFullSubscription            = testDefs.RunSubscriptionFull
	resourceAdapterIntegrationTest = resourceAdapterTest.ResourceAdapterIntegrationTest
	sendToMaintenanceRMQ           = testDefs.PostMaintenance2RMQ
)

func main() {
	const fct = "[main]"

	startServerPtr := flag.String("startServer", "", "any value here will start the server")
	testNamePtr := flag.String("test", "all", "name of the test to run, can be comma separated: all,incident,resource,maintenance,notification")
	endpoint := flag.String("endpoint", "https://pnp-api-oss.dev.cloud.ibm.com/catalog/api/info", "the url pointing to the catalog API info (staging: https://pnp-api-oss.test.cloud.ibm.com/catalog/api/info, production: https://pnp-api-oss.cloud.ibm.com/catalog/api/info,  old staging: https://api-oss-staging.bluemix.net/catalog/api/info, old development: https://api-oss-dev.bluemix.net/catalog/api/info")
	tokenPtr := flag.String("token", "", "API token to use on requests")
	logLevelPtr := flag.String("loglevel", "none", "specifies the log output to disable values can be info, warning, error, none")
	flag.Parse()
	lg.DisableLogLevel(*logLevelPtr)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println(fct, "StartServer =", *startServerPtr)

	if *startServerPtr != "" {
		log.Println(fct, "Version .mar20")
		log.Println(fct, "----------------------------")
		log.Println(fct, "PG DB - host:", config.PgHost)
		log.Println(fct, "PG DB - port:", config.PgPort)
		log.Println(fct, "PG DB - name:", config.PgDB)
		log.Println(fct, "PG DB - user:", config.PgDBUser)
		log.Println(fct, "PG DB - SSL: ", config.DBSSLMode)
		log.Println(fct, "----------------------------")
		log.Println(fct, "Connecting to Postgres database...")

		port, err := strconv.Atoi(config.PgPort)
		if err != nil {
			log.Println(fct, err.Error())
			return
		}

		const noSSL = "disable"
		if config.DBSSLMode == noSSL {
			config.Pdb, err = db.Connect(config.PgHost, port, config.PgDB, config.PgDBUser, config.PgPass, config.DBSSLMode)
		} else {
			config.Pdb, err = db.ConnectWithSSL(config.PgHost, port, config.PgDB, config.PgDBUser, config.PgPass, config.DBSSLMode, config.PgSSLRootCertFilePath)
		}
		if err != nil {
			log.Println(fct, err.Error())
		}

		// Check that the database parameters passed are valid and database connection is valid
		err = chkdbcon.CheckDBConnection(config.Pdb)
		if err != nil {
			log.Println(fct, err.Error())
			return
		}

		log.Println(fct, "Successfully connected to Postgres database")

		http.HandleFunc("/", simpleHandler)
		http.HandleFunc("/help", help)
		http.HandleFunc("/run/", runHandler)

		log.Fatal(http.ListenAndServe(":8000", nil))

		// if fileExists("ssl.crt") {
		// 	log.Fatal(http.ListenAndServeTLS(":8443", "ssl.crt", "ssl.cl", nil))
		// }
	} else {
		cat := catalog.NewCatalog(*endpoint, &rest.Server{Token: *tokenPtr})

		runTests(cat, strings.Split(*testNamePtr, ","))

		lg.ShowTestResult()
	}
}

// Page represents a web page
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) run() {
	log.Println("Running title =", p.Title)
}

func help(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "curl localhost:8000/run/?basic\n")
	fmt.Fprintf(w, "curl localhost:8000/run/?RunSubscriptionFull\n")
	fmt.Fprintf(w, "curl localhost:8000/run/?notifications\n")
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	const fct = "[runHandler]"
	log.Println(fct, "starting")

	var retVal, isSuccess string
	query := r.URL.Query()

	// curl localhost:8000/run/?basic
	_, ok := query["basic"]
	if ok {
		log.Println(fct, "Running Basic Test")
		hooksHandler(w)
		return
	}

	// curl localhost:8000/run/?RunSubscriptionFull
	_, ok = query["RunSubscriptionFull"]
	if ok {
		log.Println(fct, "Running SubscriptionFull")
		retVal = runFullSubscription()
	}

	// curl localhost:8000/run/?ResourceAdapterIntegrationTest
	_, ok = query["ResourceAdapterIntegrationTest"]
	if ok {
		log.Println(fct, "Running ResourceAdapterIntegrationTest")
		retVal = resourceAdapterIntegrationTest()
	}

	// curl localhost:8000/run/?CaseAPIIntegrationTest
	// _, ok = query["CaseAPIIntegrationTest"]
	// if ok {
	// 	log.Println(fct, "Running CaseAPIIntegrationTest")
	// 	retVal = caseAPIIntegrationTest()
	// }

	if retVal != "" {
		log.Println(fct, "Test completed:", retVal)
		if retVal == testDefs.SUCCESS_MSG {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// curl localhost:8000/run/?sendToMaintenanceRMQ
	_, ok = query["sendToMaintenanceRMQ"]
	if ok {
		log.Println(fct, "Running sendToMaintenanceRMQ")
		err := sendToMaintenanceRMQ()
		if err != nil {
			fmt.Fprintf(w, "%s", err.Error())
		}
	} else {
		fmt.Fprintf(w, "\nNothing to do")
		return
	}

	title := r.URL.Path

	p := &Page{Title: title, Body: []byte(isSuccess)}
	//p.run()
	// http.Redirect(w, r, "/view/"+title, http.StatusFound)
	t := template.New("test run")
	t, err := t.Parse(forms.RunForm)
	if err != nil {
		log.Println(fct, err.Error())
	}
	t.Execute(w, p)
}

// Returns the map of the JSON body that was sent in. Used for test purposes
func simpleHandler(w http.ResponseWriter, r *http.Request) {
	const fct = "[http.HandleFunc->simpleHandler]"
	log.Println(fct, "---------------------------------------")

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Println(fct, err.Error())
	}
	defer r.Body.Close()

	requestMsg := r.Method + " received from : " + r.URL.Query().Get("origin") +
		"\n\tUrl -> " + r.URL.String() +
		"\n\tBody : " + string(body)
	log.Println(fct, requestMsg)

	go func() {
		testDefs.Messages <- requestMsg
	}()

	var anyJSON map[string]interface{}
	if r.Method == http.MethodPost {
		if err := json.Unmarshal(body, &anyJSON); err != nil {
			log.Println(fct, "JSON unmarshal error:", err.Error())
			w.WriteHeader(http.StatusUnprocessableEntity)
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Output : %+v\n", anyJSON)
}

func runTests(cat *catalog.Catalog, tests []string) {
	runStatusTests(cat, tests)
}

func runStatusTests(cat *catalog.Catalog, tests []string) {
	const fct = "[runStatusTests]"
	log.Println(fct, "Running status tests")

	statusAPI, err := status.NewStatusAPI(cat)
	if err != nil {
		log.Println(fct, err.Error())
		return
	}

	for _, test := range tests {
		if test == "incident" || test == "all" {
			statusAPI.IncidentTest()
		}
		if test == "maintenance" || test == "all" {
			statusAPI.MaintenanceTest()
		}
		if test == "resource" || test == "all" {
			statusAPI.ResourceTest()
		}
		if test == "notification" || test == "all" {
			statusAPI.NotificationTest()
		}
	}
}

func sendWebhook(server *rest.Server, url string, body string) {
	const fct = "[sendWebhook]"
	if server == nil {
		server = &rest.Server{}
	}

	resp, err := server.Post(fct, url, body)
	if err != nil {
		log.Println(fct, err.Error())
	}
	log.Println(fct, "response:", resp)
}
