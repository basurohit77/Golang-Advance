package handlers

import (
	"log"
	"net/http"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
)

// ServeSnowCases - services /cases requests
func ServeSnowCases(res http.ResponseWriter, req *http.Request) {

	// intercept panics: print error and stacktrace
	defer api.HandlePanics(res)
	defer req.Body.Close()

	log.Print("Received ServiceNow case, but ignoring because PNP case is disabled")
	/*
		log.Print("Invoking SNow Cases")

		// post message to rabbitMQ
		httpStatus, err := shared.HandleMessage(req.Body, "case")
		if err != nil && httpStatus != http.StatusOK {
			log.Println(err)
			res.WriteHeader(httpStatus)
			return
		}
	*/

	//log.Println("successfully produced case to RabbitMQ")
}
