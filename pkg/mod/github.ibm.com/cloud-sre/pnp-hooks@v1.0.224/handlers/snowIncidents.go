package handlers

import (
	"log"
	"net/http"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

// ServeSnowIncidents - services /incidents requests
func ServeSnowIncidents(res http.ResponseWriter, req *http.Request) {

	// intercept panics: print error and stacktrace
	defer api.HandlePanics(res)
	defer req.Body.Close()

	log.Print("Invoking SNow Incidents")

	// post message to rabbitMQ
	httpStatus, err := shared.HandleMessage(req.Body, "incident")
	if err != nil && httpStatus != http.StatusOK {
		log.Println(err)
		res.WriteHeader(httpStatus)
		return
	}

	//log.Println("successfully produced incident to RabbitMQ")

}
