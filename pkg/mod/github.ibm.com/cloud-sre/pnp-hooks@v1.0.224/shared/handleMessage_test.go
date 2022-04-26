package shared

import (
	"bytes"
	"testing"
)

func TestHandleMessage(t *testing.T) {

	if err := ConnectProducerRabbitMQ(); err != nil {
		println("unable to connect to rabbitmq. err=", err)
	}

	code, err := HandleMessage(bytes.NewReader([]byte("this is a test msg")), "test")
	if code != 200 || err != nil {
		t.Fatal(err, code)
	}
}
