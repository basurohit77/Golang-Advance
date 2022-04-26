package mq

import (
	"testing"

	rabbitmqconnector "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

func TestFunctions(t *testing.T) {
	SetupMQFunctions()

	SetupMQFunctionsForUT(utNewProducerFunc, utOpenHiLevelFunc)
	SetupMQFunctionsForUT(nil, nil)
}

func utOpenHiLevelFunc(url []string, routingKey, exchangeName, exchangeType string) (iconn IConnection, err error) {
	return nil, nil
}

func utNewProducerFunc(url []string, routingKey, exchangeName, exchangeType string) *rabbitmqconnector.AMQPProducer {
	return nil
}
