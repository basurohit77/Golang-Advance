package handlers

import (
	"log"
	"net/http"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

// ServeDoctorMaintenance - services /maintenances requests
func ServeDoctorMaintenance(res http.ResponseWriter, req *http.Request) {

	if !shared.HasValidToken(req.Header.Get("Authorization")) {
		log.Println("invalid token")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	// intercept panics: print error and stacktrace
	defer api.HandlePanics(res)
	defer req.Body.Close()

	log.Print("Invoking Doctor Maintenance")

	// post message to rabbitMQ
	httpStatus, err := shared.HandleMessage(req.Body, "maintenance")
	if err != nil && httpStatus != http.StatusOK {
		log.Println(err)
		res.WriteHeader(httpStatus)
		return
	}

	//log.Println("successfully produced case to RabbitMQ")
}
