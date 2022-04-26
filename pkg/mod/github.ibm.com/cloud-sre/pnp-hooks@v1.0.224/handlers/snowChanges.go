package handlers

import (
	"log"
	"net/http"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

// ServeSnowChanges - services /changes requests
func ServeSnowChanges(res http.ResponseWriter, req *http.Request) {

	// intercept panics: print error and stacktrace
	defer api.HandlePanics(res)
	defer req.Body.Close()

	log.Print("Invoking SNow Changes")

	// post message to rabbitMQ
	httpStatus, err := shared.HandleMessage(req.Body, "maintenance")
	if err != nil && httpStatus != http.StatusOK {
		log.Println(err)
		res.WriteHeader(httpStatus)
		return
	}

	//log.Println("successfully produced maintenance to RabbitMQ")

}
