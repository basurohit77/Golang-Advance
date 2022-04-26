package mq

import (
	"fmt"
	"log"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	rabbitmqconnector "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

const (
	// RoutingKey is the key to use for routing notifications
	RoutingKey = "notification"
	// ExchangeName is the name for rabbit mq exchange
	ExchangeName = "pnp.direct"
	// ExchangeType is the type for rabbit mq exchange
	ExchangeType = "direct"
	// UnitTestingRoutingKey is a special routing key for running unit tests.
	UnitTestingRoutingKey = "UNIT_TEST_ROUTING"
)

// IConnection abstracts the connection object returned
type IConnection interface {
	Produce(msg string) error
	SendNotifications(ctx ctxt.Context, nList []datastore.NotificationInsert, msgType datastore.NotificationMsgType) (err error)
	SendCompareAndUpdateNotifications(ctx ctxt.Context, existingList, newList []datastore.NotificationInsert) error
	// Close() error
}

// Connection is an abstraction of the Message Queue connection
type Connection struct {
	Producer *rabbitmqconnector.AMQPProducer
	// Connection *amqp.Connection
	// Channel    *amqp.Channel
}

// Open will open a connection to the message queue
// NOTICE: Instead of calling this directly, please call "OpenConnection" with same signature
func open(url []string, routingKey, exchangeName, exchangeType string) (iconn IConnection, err error) {

	METHOD := "mq.Open"

	// if url == "" {
	if len(url) < 1 {
		return nil, fmt.Errorf("ERROR (%s): missing url parameter", METHOD)
	}

	conn := new(Connection)

	conn.Producer = mqNewProducer(url, routingKey, exchangeName, exchangeType)

	if conn.Producer.RoutingKey != UnitTestingRoutingKey { // check for unit testing
		// conn.Connection, conn.Channel, err = conn.Producer.Connect()
		err = conn.Producer.Connect()
	}

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// MQEncryptionEnabled controls if messages are encrypted before going to the MQ
var MQEncryptionEnabled = true

// Produce will produce a message to the queue
func (conn *Connection) Produce(msg string) error {

	var err error
	var outputMsg []byte

	if MQEncryptionEnabled {
		outputMsg, err = encryption.Encrypt(msg)
		if err != nil {
			log.Println("pnp-notifications-adapter.mq.Produce: Encryption failure: " + err.Error())
			return err
		}
	} else {
		outputMsg = []byte(msg)
	}

	if conn.Producer.RoutingKey == UnitTestingRoutingKey { // check for unit testing
		return nil
	}

	return conn.Producer.Produce(string(outputMsg))
}

// Close will orderly shutdown of the MQ connection
// func (conn *Connection) Close() error {

// 	if conn == nil {
// 		return nil
// 	}

// 	if conn.Connection != nil {
// 		conn.Connection.Close()
// 	}

// 	if conn.Channel != nil {
// 		conn.Channel.Close()
// 	}

// 	return nil
// }
