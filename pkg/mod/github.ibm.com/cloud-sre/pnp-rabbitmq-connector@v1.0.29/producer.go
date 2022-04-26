package rabbitmqconnector

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Exchange defines all attributes to create an exchange
type Exchange struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

// AMQPProducer defines all values a producer can be configured
type AMQPProducer struct {
	URL        string
	TLSConfig  *tls.Config
	RoutingKey string
	Exchange
	Mandatory bool
	Immediate bool
	amqp.Publishing
	ch   *amqp.Channel
	conn *amqp.Connection
}

var (
	unlockProducer     = make(chan struct{})
	lockProducer       = false
	producerURLs       []string
	firstCallToConnect = true
)

// NewProducer creates and returns a new producer struct(object)
func NewProducer(url []string, routingKey, exchangeName, exchangeType string) *AMQPProducer {

	producerURLs = url

	if len(url) == 0 {
		url = []string{""}
	}

	return &AMQPProducer{
		URL:        url[0],
		RoutingKey: routingKey,
		Exchange: Exchange{
			Name: exchangeName,
			Type: exchangeType,
		},
	}
}

// NewProducer creates and returns a new producer struct(object)
func NewSSLProducer(url []string, ca_cert, routingKey, exchangeName, exchangeType string) *AMQPProducer {

	producerURLs = url

	if len(url) == 0 {
		url = []string{""}
	}

	decoded_ca, err := base64.StdEncoding.DecodeString(ca_cert)
	if err != nil {
		log.Fatalln("Failed to decode root certificate!")
		return nil
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(decoded_ca))
	if !ok {
		log.Fatalln("Failed to parse root certificate!")
		return nil
	}
	tlsConfig := &tls.Config{
		RootCAs:    roots,
		MinVersion: tls.VersionTLS12,
	}

	return &AMQPProducer{
		URL:        url[0],
		TLSConfig:  tlsConfig,
		RoutingKey: routingKey,
		Exchange: Exchange{
			Name: exchangeName,
			Type: exchangeType,
		},
	}
}

// Connect stablishes a connection to a rabbitMQ server
// func (producer *AMQPProducer) Connect() (*amqp.Connection, *amqp.Channel, error) {
func (producer *AMQPProducer) Connect() error {

	conn, ch, err := connectWithRetries(producer.URL, producer.TLSConfig)
	if err != nil {
		log.Printf("failed connecting to url=%s", getHostnameFromURL(producer.URL))
		// attempt to connect to the second URL if available
		// do this only the first time the producer is initiated
		if len(producerURLs) > 1 && producerURLs[1] != producer.URL && firstCallToConnect {
			producer.URL = producerURLs[1]
			log.Printf("attempting reconnection using url=%s", getHostnameFromURL(producer.URL))
			conn, ch, err = connectWithRetries(producer.URL, producer.TLSConfig)
			if err != nil {
				// return nil, nil, err
				return err
			}
			firstCallToConnect = false
		} else {
			// return nil, nil, err
			return err
		}
	}

	producer.conn = conn
	producer.ch = ch

	firstCallToConnect = false

	// handles unhealthy connectivity to rabbitMQ server
	go func() {

		defer producer.conn.Close()
		defer producer.ch.Close()

		<-producer.conn.NotifyClose(make(chan *amqp.Error))

		// do not let any new messages be "produced" while connectivity to rabbitMQ server is unhealthy
		lockProducer = true

		log.Printf("producer connection was closed, attempting reconnection. url=%s", getHostnameFromURL(producer.URL))

		// defer producer.conn.Close()
		// defer producer.ch.Close()

		// var err error

		// attempt to re-connect to rabbitMQ server indefinitely
		for {
			// if _, _, err = producer.Connect(); err == nil {
			if err := producer.Connect(); err == nil {
				// release lock
				lockProducer = false
				unlockProducer <- struct{}{}
				break
			}
			time.Sleep(500 * time.Millisecond)

			// keep trying to re-connect switching between url1 and url2
			if len(producerURLs) > 1 {
				if producer.URL == producerURLs[0] {
					producer.URL = producerURLs[1]
					// log.Printf("attempting to connect to RabbitMQ using url=%s", getHostnameFromURL(producer.URL))
				} else {
					producer.URL = producerURLs[0]
					// log.Printf("attempting to connect to RabbitMQ using url=%s", getHostnameFromURL(producer.URL))
				}
				log.Printf("attempting to connect to RabbitMQ using url=%s", getHostnameFromURL(producer.URL))
			}
		}
	}()

	log.Println("producer successfully connected to", getHostnameFromURL(producer.URL))

	return err
	// return producer.conn, producer.ch, err
}

// UnlockProducer releases the lock when the connection to rabbitMQ is unhealthy
// it is needed by the unit tests
// func (producer *AMQPProducer) UnlockProducer() {
// 	lockProducer = false
// }

// Produce sends a message to RabbitMQ using a specific routing key
func (producer *AMQPProducer) Produce(msg string) error {

	// lock producer until connection to rabbitmq is reestablished
	// without this lock, msgs sent with an unhealthy connection will be lost if not handled by the client
	// lock is released when connectivity with rabbitMQ server is healthy
	if lockProducer {
		<-unlockProducer
	}

	// write income msg to the body of the msg that will be sent to the server
	producer.Publishing.Body = []byte(msg)

	if producer.Publishing.ContentType == "" {
		producer.Publishing.ContentType = "application/json"
	}

	//if producer.Publishing.Timestamp == (time.Time{}) {
	// timestamp each msg
	producer.Publishing.Timestamp = time.Now().UTC()
	//}

	// msg in the queue will be persistent, even after server restart
	producer.Publishing.DeliveryMode = amqp.Persistent

	// set Exchange to durable(persistent), so it wont get deleted by the server
	producer.Exchange.Durable = true

	err := producer.ch.ExchangeDeclare(
		producer.Exchange.Name,
		producer.Exchange.Type,
		producer.Exchange.Durable,
		producer.Exchange.AutoDelete,
		producer.Exchange.Internal,
		producer.Exchange.NoWait,
		producer.Exchange.Args,
	)
	if err != nil {
		return err
	}

	// sends msg to rabbitmq
	return producer.publishWithRetries()

}

// publish sends a msg to rabbitmq specifying the exchange and routing key
func (producer *AMQPProducer) publish() error {
	return producer.ch.Publish(
		producer.Exchange.Name, // exchange
		producer.RoutingKey,    // routing key
		producer.Mandatory,     // mandatory
		producer.Immediate,     // immediate
		producer.Publishing,    // msg payload and metadata
	)
}

func (producer *AMQPProducer) publishWithRetries() (err error) {
	for i := 0; i < RetryCount; i++ {
		err = producer.publish()
		if err == nil {
			return err
		}
		//fmt.Println("Retrying to publish message")
		time.Sleep(RetryWait)
	}
	return err
}

// ProduceOnce sends only one msg per connection
// it wraps the Connect and Produce functions to make things simpler when using this lib
// if you want to send several msgs re-using the same connection(better performace),
// you need to call Connect and Produce in your code
func (producer *AMQPProducer) ProduceOnce(msg string) (err error) {

	producer.conn, producer.ch, err = producer.ConnectWithoutReconnection()
	if err != nil {
		return err
	}

	defer producer.conn.Close()
	defer producer.ch.Close()

	return producer.Produce(msg)
}

// ConnectWithoutReconnection stablishes a connection to a rabbitMQ server.
// No auto-reconnection will be attempted if connection becomes unhealthy
func (producer *AMQPProducer) ConnectWithoutReconnection() (conn *amqp.Connection, ch *amqp.Channel, err error) {

	conn, ch, err = connectWithRetries(producer.URL, producer.TLSConfig)
	if err != nil {
		log.Printf("failed connecting to url=%s", getHostnameFromURL(producer.URL))
		// attempt to connect to the second URL if available
		// do this only the first time the producer is initiated
		if len(producerURLs) > 1 && producerURLs[1] != producer.URL {
			producer.URL = producerURLs[1]
			log.Printf("attempting reconnection using url=%s", getHostnameFromURL(producer.URL))
			conn, ch, err = connectWithRetries(producer.URL, producer.TLSConfig)
			if err != nil {
				return
			}
		} else {
			return
		}
	}

	log.Println("producer successfully connected to", getHostnameFromURL(producer.URL))

	return
}
