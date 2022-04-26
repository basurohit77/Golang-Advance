package rabbitmqconnector

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

var (
	// RetryCount number of retries to connect to RabbitMQ and to send(produce) msg to a queue
	RetryCount = 3
	// RetryWait number in seconds to wait between retries
	RetryWait = 5 * time.Second
)

func connectWithRetries(url string, tlsConfig *tls.Config) (conn *amqp.Connection, ch *amqp.Channel, err error) {
	for i := 0; i < RetryCount; i++ {
		conn, ch, err = connect(url, tlsConfig)
		if err == nil {
			return conn, ch, err
		}
		time.Sleep(RetryWait)
	}

	return nil, nil, err
}

func connect(url string, tlsConfig *tls.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.DialTLS(url, tlsConfig)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	return conn, ch, nil
}

func getHostnameFromURL(url string) string {
	if strings.Contains(url, "@") && strings.Contains(url, ":") {
		splitURL := strings.Split(url, "@")
		return strings.Split(splitURL[1], ":")[0]
	}
	return url
}
