package handlers

import (
	"net/http"
	"testing"

	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

func TestSnowChangesValidToken(t *testing.T) {

	if err := shared.ConnectProducerRabbitMQ(); err != nil {
		t.Error("unable to connect to rabbitmq. err=", err)
	}

	// defer shared.ProducerConn.Close()
	// defer shared.ProducerCh.Close()

	httpPost(t, "api/v1/snow/changes", "", http.StatusOK, ServeSnowChanges)

}

// func TestSnowChangesValidTokenRabbitMQErr(t *testing.T) {
// 	shared.Producer = nil

// 	httpPost(t, "api/v1/snow/changes", "", http.StatusInternalServerError, ServeSnowChanges, os.Getenv("SNOW_TOKEN"))
// }
