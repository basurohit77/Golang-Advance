package shared

import (
	"testing"

	"github.com/streadway/amqp"
)

func TestConnectProducerRabbitMQ(t *testing.T) {

	if err := ConnectProducerRabbitMQ(); err != nil {
		t.Fatal(err)
	}

}

// TestConnectionToRabbitMQ - tests the connection to RabbitMQ by starting a new consumer
func TestTestConnectionToRabbitMQ(t *testing.T) {
	if ok := TestConnectionToRabbitMQ(); !ok {
		t.Fatal(ok)
	}

	testConnectionToRabbitMQF(amqp.Delivery{})
}
