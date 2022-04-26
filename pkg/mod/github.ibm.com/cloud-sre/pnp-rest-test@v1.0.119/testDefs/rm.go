package testDefs

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	rabbitmq "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

// Post2Hooks - Sends a payload to the hooks server
// msg can be:
// incidentUpdateTime = time.Now().Format("2006-01-02 15:04:05")
// -> fmt.Sprintf(createIncidentMessageInRMQ, incidentUpdateTime))
func Post2Hooks(url string, msg string) error {
	const fct = "Post2Hooks"
	lg.Info(fct, "********** Starting **********")
	if url == "" {
		url = os.Getenv("RMQHooksDr")
	}
	serv := rest.Server{}
	serv.Token = os.Getenv("HOOK_KEY")

	// *************************************
	// Send the incident data via the hooks server

	updateTime := time.Now().Format("2006-01-02 15:04:05")
	lg.Info(fct, "updateTime = "+updateTime, "\nurl = "+url, "\nmsg = "+msg)
	resp, err := serv.Post(fct, url, msg)
	if err != nil {
		lg.Err(fct, err)
		return err
	}
	lg.Info(fct, resp.Status)
	return nil
}

// PostDr2Hooks - Sends a payload to the hooks server
// msg can be:
// incidentUpdateTime = time.Now().Format("2006-01-02 15:04:05")
// -> fmt.Sprintf(createDrMessageInRMQ, incidentUpdateTime))
func PostDr2Hooks(crn string) (string, error) {
	const fct = "PostDr2Hooks"
	lg.Info(fct, "********** Starting **********")

	url := os.Getenv("RMQHooksDr")
	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day() + 1
	millis := "-" + strconv.Itoa(int(time.Now().Unix())%1250000000)
	lg.Info(fct, "millis: "+millis)
	randomBytes := make([]byte, 7)
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	_, err := rand.Read(randomBytes)
	if err != nil {
		lg.Err(fct, err)
	}
	for i, b := range randomBytes {
		randomBytes[i] = letters[b%byte(len(letters))]
	}

	rstr := string(randomBytes) + millis
	msg := fmt.Sprintf(createDrMessageInRMQ, year, month, day, year, month, day, year, month, day, rstr, rstr, rstr)
	sourceid := "2222243-" + rstr
	log.Print(fct + ": sourceid: " + sourceid + "\t crn: " + crn)
	recordid := db.CreateNotificationRecordID("Doctor-RTC", sourceid, crn, "", "maintenance")

	return recordid, Post2Hooks(url, msg)

}
func PostBspn2Hooks(crn string) (string, error) {
	const fct = "PostBspn2Hooks"
	lg.Info(fct, "********** Starting PostBspn2Hooks **********")

	url := os.Getenv("RMQHooksBspn")

	nanos := int(time.Now().UnixNano())
	millis := nanos / 1000000
	idStr := "-" + strconv.Itoa(millis%1544660000000)
	dateStr := strconv.Itoa(millis)
	msg := fmt.Sprintf(createBSPNViaHooks, dateStr, idStr, dateStr)
	recordId := db.CreateNotificationRecordID("servicenow"+"BSPN", idStr, crn, "", "incident")
	log.Print(fct + ": string to record id: " + "servicenow" + "BSPN" + idStr + crn)
	return recordId, Post2Hooks(url, msg)

}

// msg can be:
// maintenanceUpdateTime := time.Now().Format(time.RFC3339)
// -> fmt.Sprintf(createMaintenanceMessageinRMQ, maintenanceUpdateTime))
func PostDrMaintenance2RMQ(msg string) error {
	// *************************************
	// Send the maintenance data directly to RMQ
	const fct = "PostDrMaintenance2RMQ"
	lg.Info(fct, "********** Starting **********")
	var RMQUrls []string
	var p1 *rabbitmq.AMQPProducer
	if isTargetMessagesForRabbitMQ() {
		RMQUrls = append(RMQUrls, RMQAMQPSEndpoint)
		p1 = rabbitmq.NewSSLProducer(RMQUrls, RMQTLSCert, RMQRoutingKeyMaintenance, exchangeName, exchangeType)
		lg.Info(fct, strings.Split(RMQAMQPSEndpoint, "@")[1][:20], RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	} else {
		RMQUrls = append(RMQUrls, RMQUrl)
		p1 = rabbitmq.NewProducer(RMQUrls, RMQRoutingKeyMaintenance, exchangeName, exchangeType)
		lg.Info(fct, strings.Split(RMQUrl, "@")[1][:20], RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	}
	//maintenanceUpdateTime := time.Now().Format(time.RFC3339)

	encMaintMsg, err := encryption.Encrypt(msg)
	if err != nil {
		return err
	}

	err = p1.ProduceOnce(string(encMaintMsg))
	if err != nil {
		log.Fatalln(err)
	}
	lg.Info(fct, "Posted maintenance message to Rabbit MQ")
	return nil
}

//
func PostStatus2RMQ(msg string) error {
	const fct = "PostStatus2RMQ"
	lg.Info(fct, "********** Starting **********")
	// *************************************
	// Send the status data directly to RMQ

	lg.Info(fct, strings.Split(RMQUrl, "@")[1][:20], RMQRoutingKeyStatus, exchangeName, exchangeType)
	RMQUrls := []string{RMQUrl}
	p2 := rabbitmq.NewProducer(RMQUrls, RMQRoutingKeyStatus, exchangeName, exchangeType)

	// Encryption
	//statusUpdateTime := time.Now().Format(time.RFC3339)
	encStatusMsg, err := encryption.Encrypt(msg)
	if err != nil {
		return err
	}

	err = p2.ProduceOnce(string(encStatusMsg))
	if err != nil {
		log.Fatalln(err)
	}
	lg.Info(fct, "Posted status message to Rabbit MQ")

	return nil
}

func PostMsg2RMQ(msg string, routingKey string) {
	const FCT = "Test_Encryption: "
	//msg := `{"id":"accesstrail","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"accesstrail","creationTime":"2015-12-11T20:01:39Z","updateTime":"2018-08-22T20:46:03.8Z","deployments":[{"id":"free-au-syd","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:au-syd::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:au-syd::::","creationTime":"2018-01-18T05:53:47Z","updateTime":"2018-03-09T06:32:48.007Z","parent":true},{"id":"free-eu-de","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:eu-de::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:eu-de::::","creationTime":"2017-03-21T03:25:34Z","updateTime":"2018-03-09T06:32:48.261Z","parent":true},{"id":"lite-eu-gb","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:eu-gb::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:eu-gb::::","creationTime":"2018-02-01T06:55:49Z","updateTime":"2018-03-09T06:32:48.209Z","parent":true},{"id":"free-us-south","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:us-south::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:us-south::::","creationTime":"2015-12-11T20:01:39Z","updateTime":"2018-03-09T06:32:47.929Z","parent":true}]}`
	if msg == "" {
		msg = `{"id":"accesstrail","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus"],"source":"globalCatalog","sourceId":"accesstrail","creationTime":"2015-12-11T20:01:39Z","updateTime":"2018-08-22T20:46:03.8Z","deployments":[{"id":"free-au-syd","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:au-syd::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:au-syd::::","creationTime":"2018-01-18T05:53:47Z","updateTime":"2018-03-09T06:32:48.007Z","parent":true},{"id":"free-eu-de","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:eu-de::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:eu-de::::","creationTime":"2017-03-21T03:25:34Z","updateTime":"2018-03-09T06:32:48.261Z","parent":true},{"id":"lite-eu-gb","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:eu-gb::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:eu-gb::::","creationTime":"2018-02-01T06:55:49Z","updateTime":"2018-03-09T06:32:48.209Z","parent":true},{"id":"free-us-south","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:us-south::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:us-south::::","creationTime":"2015-12-11T20:01:39Z","updateTime":"2018-03-09T06:32:47.929Z","parent":true}]}`
	}

	encryptedData, _ := encryption.Encrypt(msg)
	mqP := createProducer(routingKey)
	if err := messageProducer(mqP, string(encryptedData)); err != nil {
		log.Fatal(err)
	}
}
func createProducer(routingKey string) (mqP *rabbitmq.AMQPProducer) {
	const FCT = "createProducer: "

	var urls []string
	if isTargetMessagesForRabbitMQ() {
		mqUrl := os.Getenv("RABBITMQ_AMQPS_ENDPOINT")
		mqCert := os.Getenv("RABBITMQ_TLS_CERT")
		urls = append(urls, mqUrl)
		mqP = rabbitmq.NewSSLProducer(urls, mqCert, routingKey, "pnp.direct", "direct")
	} else {
		mqUrl := os.Getenv("NQ_URL")
		urls = append(urls, mqUrl)
		mqP = rabbitmq.NewProducer(urls, routingKey, "pnp.direct", "direct")
	}

	log.Print(urls)
	var err error
	// conn, ch, err = mqP.Connect()
	err = mqP.Connect()
	if err != nil {
		log.Fatalln(FCT + err.Error())
	}
	// return mqP, conn, ch
	return mqP
}

func messageProducer(producer *rabbitmq.AMQPProducer, msg string) (err error) {

	if len(msg) > 37 {
		log.Println("RMQ Message : ", msg[:37])
	} else {
		log.Println("RMQ Message : ", msg)
	}

	producer.Publishing.MessageId = "OSS Catalog"

	encryptedData, err := encryption.Encrypt(msg)
	if err != nil {
		log.Println(err)

		return err
	}

	err = producer.Produce(string(encryptedData))
	if err != nil {
		log.Println(err)

		return err
	}

	return err
}

// msg can be:
// statusUpdateTime := time.Now().Format(time.RFC3339)
// -> fmt.Sprintf(createStatusMessageInRMQ, statusUpdateTime)
func postBasic2RMQ(w http.ResponseWriter) error {
	const fct = "postBasic2RMQ"
	lg.Info(fct, "********** Starting **********")
	time.Sleep(1 * time.Second)
	serv := rest.Server{}
	serv.Token = os.Getenv("HOOK_KEY")

	// *************************************
	// Send the incident data via the hooks server
	time.Sleep(1 * time.Second)
	incidentUpdateTime = time.Now().Format("2006-01-02 15:04:05")
	log.Println(fct, "incidentUpdateTime = "+incidentUpdateTime)
	_, err := serv.Post(fct, RMQHooksIncident, fmt.Sprintf(createIncidentMessageInRMQ, incidentUpdateTime))
	if err != nil {
		log.Println(fct, "Error = "+err.Error())
		return err
	}

	// *************************************
	// Send the maintenance data directly to RMQ
	time.Sleep(1 * time.Second)
	var RMQUrls []string
	var p1 *rabbitmq.AMQPProducer
	if isTargetMessagesForRabbitMQ() {
		RMQUrls = append(RMQUrls, RMQAMQPSEndpoint)
		p1 = rabbitmq.NewSSLProducer(RMQUrls, RMQTLSCert, RMQRoutingKeyMaintenance, exchangeName, exchangeType)
		lg.Info(fct, strings.Split(RMQAMQPSEndpoint, "@")[1][:20], RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	} else {
		RMQUrls = append(RMQUrls, RMQUrl)
		p1 = rabbitmq.NewProducer(RMQUrls, RMQRoutingKeyMaintenance, exchangeName, exchangeType)
		lg.Info(fct, strings.Split(RMQUrl, "@")[1][:20], RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	}

	maintenanceUpdateTime := time.Now().Format(time.RFC3339)
	encMaintMsg, err := encryption.Encrypt(fmt.Sprintf(CreateMaintenanceMessageinRMQ, maintenanceUpdateTime))
	if err != nil {
		return err
	}

	err = p1.ProduceOnce(string(encMaintMsg))
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(fct, "Posted maintenance message to Rabbit MQ")

	// *************************************
	// Send the status data directly to RMQ
	time.Sleep(1 * time.Second)
	lg.Info(fct, strings.Split(RMQUrl, "@")[1][:20], RMQRoutingKeyStatus, exchangeName, exchangeType)
	var p2 *rabbitmq.AMQPProducer
	if isTargetMessagesForRabbitMQ() {
		p1 = rabbitmq.NewSSLProducer(RMQUrls, RMQTLSCert, RMQRoutingKeyStatus, exchangeName, exchangeType)
		lg.Info(fct, strings.Split(RMQAMQPSEndpoint, "@")[1][:20], RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	} else {
		p1 = rabbitmq.NewProducer(RMQUrls, RMQRoutingKeyStatus, exchangeName, exchangeType)
		lg.Info(fct, strings.Split(RMQUrl, "@")[1][:20], RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	}

	// Encryption
	statusUpdateTime := time.Now().Format(time.RFC3339)
	encStatusMsg, err := encryption.Encrypt(fmt.Sprintf(createStatusMessageInRMQ, statusUpdateTime))
	if err != nil {
		return err
	}

	err = p2.ProduceOnce(string(encStatusMsg))
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(fct, "Posted status message to Rabbit MQ")

	return nil
}
