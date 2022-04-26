package shared

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
)

// HandleMessage reads msg from request, and sends it to RabbitMQ using the specified routing key.
// returns an HTTP status and an error
func HandleMessage(reqBody io.Reader, routingKey string) (int, error) {

	data, err := ioutil.ReadAll(reqBody)
	if err != nil {
		return http.StatusBadRequest, err
	}

	log.Print("Raw message: ", string(api.RedactAttributes(data))) // set as exclusion rule in LogDNA

	Producer.RoutingKey = routingKey

	rawMessage := string(data)

	encryptedData, err := encryption.Encrypt(rawMessage)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = Producer.Produce(string(encryptedData))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
