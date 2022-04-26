package mq

import (
	rabbitmqconnector "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

// The purpose of this file is to map MQ functions to actual implementations.  This is primarily helpful to enable unit tests.

// NewProducerFunc is mock-able for UT purposes
type NewProducerFunc func(url []string, routingKey, exchangeName, exchangeType string) *rabbitmqconnector.AMQPProducer

// OpenHiLevelFunc is a high level declare for the function creating MQ connections
type OpenHiLevelFunc func(url []string, routingKey, exchangeName, exchangeType string) (iconn IConnection, err error)

var mqNewProducer NewProducerFunc

// OpenConnection should be used to open the connection with the MQ
var OpenConnection OpenHiLevelFunc

// SetupMQFunctions assigns functions to the MQ functions.
func SetupMQFunctions() {
	mqNewProducer = rabbitmqconnector.NewProducer
	OpenConnection = open
}

// SetupMQFunctionsForUT assigns functions to the MQ functions.
func SetupMQFunctionsForUT(f1 NewProducerFunc, f2 OpenHiLevelFunc) {
	mqNewProducer = f1
	if mqNewProducer == nil {
		mqNewProducer = rabbitmqconnector.NewProducer
	}

	OpenConnection = f2
	if OpenConnection == nil {
		OpenConnection = open
	}
}
