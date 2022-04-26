package rabbitmqconnector

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"github.com/streadway/amqp"
	"log"
	"strings"
	"time"
)

// AMQPConsumer defines all values a consumer can be configured
type AMQPConsumer struct {
	URL       string
	TLSConfig *tls.Config
	Q         []Q
	Name      string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
	ch        *amqp.Channel
	conn      *amqp.Connection
	Exchange
}

// Q defines all values of a queue in RabbitMQ
type Q struct {
	Name       string
	Key        string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type msgDelivery func(amqp.Delivery)

var (
	consumerName string
	consumerURLs []string
	// firstCallToConsume = true
)

func NewConsumer(url []string, qKey, exchangeName string) *AMQPConsumer {

	// split queues and keys
	qkall := strings.Split(qKey, ",")
	// queue config
	q := make([]Q, len(qkall))

	for k, v := range qkall {
		qk := strings.Split(v, ":")
		// first argument[0] should be the queue name
		q[k].Name = qk[0]
		// second argument[1] should be the routing key
		q[k].Key = qk[1]
	}

	consumerURLs = url

	return &AMQPConsumer{
		URL: url[0],
		Q:   q,
		Exchange: Exchange{
			Name: exchangeName,
		},
	}
}

func NewSSLConsumer(url []string, ca_cert, qKey, exchangeName string) *AMQPConsumer {

	// split queues and keys
	qkall := strings.Split(qKey, ",")
	// queue config
	q := make([]Q, len(qkall))

	for k, v := range qkall {
		qk := strings.Split(v, ":")
		// first argument[0] should be the queue name
		q[k].Name = qk[0]
		// second argument[1] should be the routing key
		q[k].Key = qk[1]
	}

	consumerURLs = url
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

	return &AMQPConsumer{
		URL:       url[0],
		TLSConfig: tlsConfig,
		Q:         q,
		Exchange: Exchange{
			Name: exchangeName,
		},
	}
}

// Consume sets up connectivity with RabbitMQ server and starts reading messages from 1 or more queues
func (consumer *AMQPConsumer) Consume(f msgDelivery) (err error) {

	// set up connection and channel to rabbitmq
	consumer.conn, consumer.ch, err = connectWithRetries(consumer.URL, consumer.TLSConfig)
	if err != nil {
		log.Printf("failed connecting to url=%s", getHostnameFromURL(consumer.URL))
		// attempt to connect to the second URL if available
		// do this only the first time Consume is called
		if len(consumerURLs) > 1 && consumerURLs[1] != consumer.URL {
			// firstCallToConsume = false
			consumer.URL = consumerURLs[1]
			log.Printf("attempting reconnection using url=%s", getHostnameFromURL(consumer.URL))
			consumer.conn, consumer.ch, err = connectWithRetries(consumer.URL, consumer.TLSConfig)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// firstCallToConsume = false

	defer consumer.conn.Close()
	defer consumer.ch.Close()

	reconnerr := make(chan error)

	go func() {
		<-consumer.conn.NotifyClose(make(chan *amqp.Error))

		log.Printf("consumer connection was closed, trying to reconnect. url=%s", getHostnameFromURL(consumer.URL))

		defer consumer.ch.Close()
		defer consumer.conn.Close()

		// attempt to re-connect to rabbitMQ server indefinitely
		for {
			// if unable to reconnect after 60s, raise an error and exit func
			// if retry > (60 * time.Second) {
			// 	reconnerr <- errors.New("unable to reconnect")
			// }

			if err := consumer.Consume(f); err == nil {
				break
			}
			time.Sleep(500 * time.Millisecond)

			// keep trying to re-connect switching between url1 and url2
			if len(consumerURLs) > 1 {
				if consumer.URL == consumerURLs[0] {
					consumer.URL = consumerURLs[1]
				} else {
					consumer.URL = consumerURLs[0]
				}
				log.Printf("attempting to connect to RabbitMQ using url=%s", getHostnameFromURL(consumer.URL))
			}

		}
	}()

	// Qos helps improve performance when AutoAck is off
	// We want to receive at least X msgs at any time
	// X = number of queues we consume from
	// err = consumer.ch.Qos(len(consumer.Q), 0, false)
	// if err != nil {
	// 	return err
	// }

	if consumer.Name != "" {
		consumerName = consumer.Name
	}

	for k := range consumer.Q {
		err := consumer.setupQueue(k, f)
		if err != nil {
			return err
		}
	}

	log.Println("consumer successfully connected to", getHostnameFromURL(consumer.URL))

	// do not let function to finish, as it would clean up(kill)
	// the goroutines created to consume data from the queue(s)
	return <-reconnerr

}

func (consumer *AMQPConsumer) setupQueue(idx int, f msgDelivery) error {

	// durable queue survives server restart
	consumer.Q[idx].Durable = true

	_, err := consumer.ch.QueueDeclare(
		consumer.Q[idx].Name,       // queue name
		consumer.Q[idx].Durable,    // durable
		consumer.Q[idx].AutoDelete, // delete when unused
		consumer.Q[idx].Exclusive,  // exclusive
		consumer.Q[idx].NoWait,     // no-wait
		consumer.Q[idx].Args,       // arguments
	)
	if err != nil {
		return err
	}

	err = consumer.ch.QueueBind(
		consumer.Q[idx].Name,   // queue name
		consumer.Q[idx].Key,    // routing key
		consumer.Exchange.Name, // exchange name
		consumer.Q[idx].NoWait, // no wait
		consumer.Q[idx].Args,   // args
	)
	if err != nil {
		return err
	}

	return consumer.consumeFromQueue(idx, f)
}

func (consumer *AMQPConsumer) consumeFromQueue(idx int, f msgDelivery) error {

	if consumer.Name != "" {
		consumer.Name = consumerName + "_" + consumer.Q[idx].Name
	}

	msgs, err := consumer.ch.Consume(
		consumer.Q[idx].Name, // queue
		consumer.Name,        // consumer
		consumer.AutoAck,     // auto-ack
		consumer.Exclusive,   // exclusive
		consumer.NoLocal,     // no-local
		consumer.NoWait,      // no-wait
		consumer.Args,        // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			f(msg)
		}
	}()

	return err
}

// ConsumeWithoutReconnection sets up connectivity with RabbitMQ server and starts reading messages from 1 or more queues.
// No automatic reconnection is done when using this function
func (consumer *AMQPConsumer) ConsumeWithoutReconnection(f msgDelivery) (*amqp.Connection, *amqp.Channel, error) {
	var err error
	// set up connection and channel to rabbitmq
	consumer.conn, consumer.ch, err = connectWithRetries(consumer.URL, consumer.TLSConfig)
	if err != nil {
		log.Printf("failed connecting to url=%s", getHostnameFromURL(consumer.URL))
		if len(consumerURLs) > 1 && consumerURLs[1] != consumer.URL {
			consumer.URL = consumerURLs[1]
			log.Printf("attempting reconnection using url=%s", getHostnameFromURL(consumer.URL))
			consumer.conn, consumer.ch, err = connectWithRetries(consumer.URL, consumer.TLSConfig)
			if err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	}

	if consumer.Name != "" {
		consumerName = consumer.Name
	}

	for k := range consumer.Q {
		err := consumer.setupQueue(k, f)
		if err != nil {
			return nil, nil, err
		}
	}

	log.Println("consumer successfully connected to", getHostnameFromURL(consumer.URL))

	return consumer.conn, consumer.ch, err

}
