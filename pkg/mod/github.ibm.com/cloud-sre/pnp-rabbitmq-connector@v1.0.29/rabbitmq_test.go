package rabbitmqconnector

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	//rabbitmq "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"

	"github.com/streadway/amqp"
)

type message struct {
	msgID   string
	msgBody string
}
type producerData struct {
	routingKey string
	msg        message
}

var (
	url          = []string{"amqp://guest:guest@localhost:5672"}
	exchangeName = "pnp.direct"
	exchangeType = "direct"
	rountingKeys = []string{"single", "multiple"}
	testData     = [2]producerData{}
)

func prepareTestData() {
	for i := 0; i < len(rountingKeys); i++ {
		testData[i] = producerData{routingKey: rountingKeys[i], msg: message{msgID: "msgID " + rountingKeys[i], msgBody: "msgBody " + rountingKeys[i]}}
	}
}

func initProducer(t *testing.T) {
	t.Logf("Init producer to create exchange %s", exchangeName)
	prepareTestData()

	if os.Getenv("NQ_URL") != "" {
		url[0] = os.Getenv("NQ_URL")
	}
	p := NewProducer(url, "", exchangeName, exchangeType)
	// conn, ch, err := p.Connect()
	err := p.Connect()
	if err != nil {
		t.Errorf("error was not expected: %s", err)
	}

	// defer conn.Close()
	// defer ch.Close()

	//test msg
	for _, td := range testData {
		p.RoutingKey = td.routingKey
		p.Publishing.MessageId = td.msg.msgID
		err = p.Produce(td.msg.msgBody)
		if err != nil {
			t.Errorf("error was not expected: %s", err)
		} else {
			t.Logf("Produced message for routingKey %s", td.routingKey)
		}
	}
}

func consumeMsg(qKey string, t *testing.T) {
	t.Logf("Create queues %s", qKey)
	c := NewConsumer(url, qKey, exchangeName)
	c.AutoAck = true
	err := c.Consume(f)
	if err != nil {
		t.Errorf("error was not expected: %s", err)
	} else {
		t.Logf("here.........")
	}
}


func Test_prepare(t *testing.T) {
	initProducer(t)
	t.Logf("Init consumer to create queues")
	for _, rk := range rountingKeys {
		var qKey = "test." + rk + ":" + rk
		go consumeMsg(qKey, t)
		time.Sleep(500 * time.Millisecond)
	}
}

func Test_Produce(t *testing.T) {
	p := NewProducer(url, "", exchangeName, exchangeType)
	// conn, ch, err := p.Connect()
	err := p.Connect()
	if err != nil {
		t.Errorf("error was not expected: %s", err)
	}

	// defer conn.Close()
	// defer ch.Close()

	//test msg
	for _, td := range testData {
		t.Run(td.routingKey, func(t *testing.T) {
			p.RoutingKey = td.routingKey
			p.Publishing.MessageId = td.msg.msgID
			err = p.Produce(td.msg.msgBody)
			if err != nil {
				t.Errorf("error was not expected: %s", err)
			} else {
				t.Logf("Produced message %s", td.msg.msgID)
			}
		})
	}
}
func Test_ProduceOnce(t *testing.T) {
	p := NewProducer(url, "", exchangeName, exchangeType)
	p.RoutingKey = testData[0].routingKey
	p.Publishing.MessageId = testData[0].msg.msgID
	err := p.ProduceOnce(testData[0].msg.msgBody)
	if err != nil {
		t.Errorf("error was not expected: %s", err)
	} else {
		t.Logf("Produced message %s", testData[0].msg.msgID)
	}

	//connection should be closed
	// p.RoutingKey = testData[1].routingKey
	// p.Publishing.MessageId = testData[1].msg.msgID
	// err = p.Produce(testData[1].msg.msgBody)
	// if err == nil {
	// 	t.Errorf("Expecting connection error.")
	// } else {
	// 	t.Logf("Connection error: %s", err)
	// }
}

func Test_NewConsumer(t *testing.T) {
	var qKeys = ""
	for _, rk := range rountingKeys {
		qKeys += "test." + rk + ":" + rk + ","
	}
	qKeys = strings.TrimSuffix(qKeys, ",")
	go consumeMsg(qKeys, t)
}

func Test_ConsumeWithoutReconnection(t *testing.T) {
	initProducer(t)
	t.Logf("Init consumer to create queues")
	for _, rk := range rountingKeys {
		var qKey = "test." + rk + ":" + rk
		t.Logf("Create queues %s", qKey)
		c := NewConsumer(url, qKey, exchangeName)
		c.AutoAck = true
		_, _, err := c.ConsumeWithoutReconnection(f)
		if err != nil {
			t.Errorf("error was not expected: %s", err)
		} else {
			t.Logf("here.........")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func f(msg amqp.Delivery) {
	log.Printf("Verify result")
	log.Printf("msg routing key: %s", msg.RoutingKey)
	log.Printf("msg : %s - %s: %s", msg.Timestamp, msg.MessageId, string(msg.Body))
}
