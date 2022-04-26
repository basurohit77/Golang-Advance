package handlers

import (
	"net/http"
	"testing"

	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

func TestDoctorMaintenanceInvalidToken(t *testing.T) {

	if err := shared.ConnectProducerRabbitMQ(); err != nil {
		t.Errorf("unable to connect to rabbitmq. err=%s", err)
	}

	// defer shared.ProducerConn.Close()
	// defer shared.ProducerCh.Close()

	httpPost(t, "api/v1/doctor/maintenances", "", http.StatusUnauthorized, ServeDoctorMaintenance)

}

// func TestDoctorMaintenanceValidTokenRabbitMQErr(t *testing.T) {
// 	// t.Log(shared.ProducerCh.Close())

// 	shared.Producer.UnlockProducer()

// 	httpPost(t, "api/v1/doctor/maintenances", "", http.StatusInternalServerError, ServeDoctorMaintenance, os.Getenv("SNOW_TOKEN"))

// }
