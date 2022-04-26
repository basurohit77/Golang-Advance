package mq

import (
	"testing"

	rabbitmqconnector "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

func TestAPI(t *testing.T) {

	SetupMQFunctionsForUT(myUTNewProducerFunc, myUTOpenHiLevelFunc)
	open([]string{"url"}, "routingKey", "exchangeName", "exchangeType")
	open([]string{""}, "routingKey", "exchangeName", "exchangeType")
}

func TestAPIError(t *testing.T) {

	SetupMQFunctionsForUT(myUTNewProducerErrorFunc, myUTOpenHiLevelFunc)
	open([]string{"url"}, "routingKey", "exchangeName", "exchangeType")
}

// func myUTNewProducerFunc(url, routingKey, exchangeName, exchangeType string) *rabbitmqconnector.AMQPProducer {
func myUTNewProducerFunc(url []string, routingKey, exchangeName, exchangeType string) *rabbitmqconnector.AMQPProducer {
	return &rabbitmqconnector.AMQPProducer{RoutingKey: UnitTestingRoutingKey}
}

func myUTNewProducerErrorFunc(url []string, routingKey, exchangeName, exchangeType string) *rabbitmqconnector.AMQPProducer {
	return &rabbitmqconnector.AMQPProducer{RoutingKey: "NOTFOUND"}
}

// OpenHiLevelFunc is a high level declare for the function creating MQ connections
func myUTOpenHiLevelFunc(url []string, routingKey, exchangeName, exchangeType string) (iconn IConnection, err error) {
	return nil, nil
}

func TestMQ(t *testing.T) {
	conn := new(Connection)
	conn.Producer = myUTNewProducerFunc([]string{"url"}, "routingKey", "exchangeName", "exchangeType")
	conn.Produce("Hello World")

	MQEncryptionEnabled = false
	conn.Produce("Hello World")
}

// func TestMQClose(t *testing.T) {
// 	conn := new(Connection)
// 	// conn.Close()
// 	conn = nil
// conn.Close()
// }
