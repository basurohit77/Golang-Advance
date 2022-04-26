package shared

import (
	"log"
	"os"

	"github.com/streadway/amqp"
	rabbitmq "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

var (
	// Producer holds the producer implementation, in which you can "produce" messages to RabbitMQ
	Producer *rabbitmq.AMQPProducer
	// RabbitMQURL holds the protocol, username, password and hostname of a RabbitMQ server
	RabbitMQURL = os.Getenv("RABBITMQ_URL")
	// RabbitMQURL2 is the second URL to the same RabbitMQ server, as a backup in case the first one is not working
	RabbitMQURL2 = os.Getenv("RABBITMQ_URL2")
	// RabbitMQEnableMessages determines whether Compose or Messages for RabbitMQ is used. If true, Messages for RabbitMQ is used.
	RabbitMQEnableMessages = os.Getenv("RABBITMQ_ENABLE_MESSAGES")
	// RabbitMQAMQPSEndpoint hold the protocal, username, password and hostname of the Messages for RabbitMQ server
	RabbitMQAMQPSEndpoint = os.Getenv("RABBITMQ_AMQPS_ENDPOINT")
	// RabbitMQTLSCert is the certificate used to connect to Messages for RabbitMQ
	RabbitMQTLSCert = os.Getenv("RABBITMQ_TLS_CERT")
	// RabbitMQRoutingKey is the routing key the producer will use to set up connectivity to RabbitMQ
	RabbitMQRoutingKey = "incident"
	// RabbitMQExchangeName is the exchange name for RabbitMQ
	RabbitMQExchangeName = os.Getenv("RABBITMQ_EXCHANGE_NAME")
	// RabbitMQExchangeType is the exchange type for RabbitMQ
	RabbitMQExchangeType = os.Getenv("RABBITMQ_EXCHANGE_TYPE")
	// RabbitMQTestQKey is used when testing connections to RabbitMQ
	RabbitMQTestQKey = os.Getenv("RABBITMQ_TEST_QKEY")
)

// ConnectProducerRabbitMQ sets up connectivity to RabbitMQ to produce messages
func ConnectProducerRabbitMQ() error {

	var urls []string
	if isTargetMessagesForRabbitMQ() {
		urls = append(urls, RabbitMQAMQPSEndpoint)
	} else {
		if RabbitMQURL != "" {
			urls = append(urls, RabbitMQURL)
		}
		if RabbitMQURL2 != "" {
			urls = append(urls, RabbitMQURL2)
		}
	}

	if isTargetMessagesForRabbitMQ() {
		Producer = rabbitmq.NewSSLProducer(urls, RabbitMQTLSCert, RabbitMQRoutingKey, RabbitMQExchangeName, RabbitMQExchangeType)
	} else {
		Producer = rabbitmq.NewProducer(urls, RabbitMQRoutingKey, RabbitMQExchangeName, RabbitMQExchangeType)
	}

	var err error

	err = Producer.Connect()
	if err != nil {
		log.Println("Unable to connect to rabbitmq")
		return err
	}

	return err

}

// TestConnectionToRabbitMQ - tests the connection to RabbitMQ by starting a new consumer
func TestConnectionToRabbitMQ() bool {
	var urls []string
	if isTargetMessagesForRabbitMQ() {
		urls = append(urls, RabbitMQAMQPSEndpoint)
	} else {
		if RabbitMQURL != "" {
			urls = append(urls, RabbitMQURL)
		}
		if RabbitMQURL2 != "" {
			urls = append(urls, RabbitMQURL2)
		}
	}

	var c *rabbitmq.AMQPConsumer
	if isTargetMessagesForRabbitMQ() {
		c = rabbitmq.NewSSLConsumer(urls, RabbitMQTLSCert, RabbitMQTestQKey, RabbitMQExchangeName)
	} else {
		c = rabbitmq.NewConsumer(urls, RabbitMQTestQKey, RabbitMQExchangeName)
	}

	c.Name = "pnp-hooks"
	c.AutoAck = true

	conn, ch, err := c.ConsumeWithoutReconnection(testConnectionToRabbitMQF)
	if err != nil {
		log.Println(err)
		return false
	}

	defer conn.Close()
	defer ch.Close()

	return true
}

func testConnectionToRabbitMQF(msg amqp.Delivery) {
	// nothing to do
}

func isTargetMessagesForRabbitMQ() bool {
	return RabbitMQEnableMessages == "true"
}
