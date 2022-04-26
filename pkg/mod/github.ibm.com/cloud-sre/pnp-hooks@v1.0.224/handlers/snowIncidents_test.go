package handlers

import (
	"net/http"
	"testing"

	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

func TestSnowIncidentsValidToken(t *testing.T) {

	if err := shared.ConnectProducerRabbitMQ(); err != nil {
		t.Errorf("unable to connect to rabbitmq. err=%s", err)
	}

	// defer shared.ProducerConn.Close()
	// defer shared.ProducerCh.Close()

	httpPost(t, "api/v1/snow/incidents", "", http.StatusOK, ServeSnowIncidents)

}

// func TestSnowIncidentsValidTokenRabbitMQErr(t *testing.T) {

// 	shared.Producer.UnlockProducer()

// 	httpPost(t, "api/v1/snow/incidents", "", http.StatusInternalServerError, ServeSnowIncidents, os.Getenv("SNOW_TOKEN"))

// }
