package liveness

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/savaki/jq" //Apache-2.0 License
	"github.ibm.com/cloud-sre/oss-globals/tlog"
)

// getQName Parse the queue names from the environment variable named nqQKey set as string value at the values.yaml
// lstQueue has the format as incident.nq2ds:incident where queue name 'incident.nq2ds' and :incident is the object need to remove the object part
// if lstQueue is null it will use the value set as default
func getQName(lstQueue string, dftLstQueue []string) []string {
	var qNames []string
	//if nqQKey is not set will use the local variable nq2dsQueues
	if lstQueue == "" {
		return dftLstQueue
	}
	qNames = strings.Split(lstQueue, ",")
	for i, q := range qNames {
		if strings.Contains(q, ":") {
			qNames[i] = q[:strings.Index(q, ":")]
		}
	}
	return qNames

}

// parseRMQEndPoint uses the rabbit endpoint set at each region as secret RABBITMQ_AMQPS_ENDPOINT the endpoint is formatted as
// TransportProtocolUser:pwd@host such as amqps://ibmxxxx:pwd12312@0023:port
// ampEndPoint rabbitMQ endpoint
func parseRMQEndPoint(ampEndPoint string) (user string, pwd string, host string) {
	ampEndPoint = ampEndPoint[len("amqps://"):] //removes the transport protocol
	user = ampEndPoint[0:strings.Index(ampEndPoint, ":")]
	strAux := ampEndPoint[strings.Index(ampEndPoint, ":")+1:]
	pwd = strAux[0:strings.Index(strAux, "@")]
	host = strAux[strings.Index(strAux, "@")+1:]
	return user, pwd, host
}

// RabbitMQHealthz does an http.Get to a rabbit host set at the region_deployment.yaml file as RABBITMQ_HOST
// if status is 200 then it is alive otherwise will return an error
// c http channel
// rabbitHost rabbitMQ Host
func RabbitMQHealthz(c *http.Client, rabbitHost string) error {
	log.Println(tlog.Log()+"Starting probe for rabbitMQ host:", rabbitHost)

	if rabbitHost == "" {
		return errors.New("Invalid host name:" + rabbitHost)
	}
	resp, err := c.Get(rabbitHost)
	if err != nil {
		return errors.New(tlog.Log() + "Failed to ping rabbitMQ host:" + rabbitHost + " error:" + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(tlog.Log() + "expecting status code of 200 got " + strconv.Itoa(resp.StatusCode) + " error:" + err.Error())
	}
	log.Println(tlog.Log()+"Successfully pinged rabbitMQ host: ", rabbitHost)
	return nil
}

// RabbitMQSize check if the messages in the queue are less than the threshold set at the values.yaml file as RABBITMQ_THRESHOLD is any of messges the listed queues
// is grater than the threshold in will set an error to liveness to ask a pod to restart
// client http channel
// rabbitAMEndPoint rabbitMQ endpoint
// rabbitHost rabbitMQ Host
// lstQueue has the format as incident.nq2ds:incident where queue name 'incident.nq2ds' and :incident is the object need to remove the object part
// dftLstQueue use in case lstQueue is empty
// qMaxSizeThresholdAllow Max number of messages allow in the queue before to restart a pod
func RabbitMQSize(client *http.Client, rabbitAMEndPoint string, rabbitHost string, lstQueue string, dftLstQueue []string, qMaxSizeThresholdAllow string) error {
	log.Println(tlog.Log() + "Starting probe for rabbitMQ queue sizes")

	user, pwd, _ := parseRMQEndPoint(rabbitAMEndPoint)
	for _, q := range getQName(lstQueue, dftLstQueue) {
		if strings.Contains(q, "case") {
			//cases are not longer in used
			continue
		}
		url := rabbitHost + "api/queues/%2F/" + q
		if rabbitHost[len(rabbitHost)-1:] != "/" { //Check if the host contains / at the end
			url = rabbitHost + "/api/queues/%2F/" + q
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println(tlog.Log() + " error:" + err.Error())
			continue
		}
		req.SetBasicAuth(user, pwd)
		req.Header.Set("Accept", "application/json")
		if err != nil {
			log.Println(tlog.Log() + " error:" + err.Error())
			continue
		}
		response, err := client.Do(req)
		if err != nil {
			log.Printf(tlog.Log(), "%s", err)
		}

		if response.StatusCode != http.StatusOK {
			log.Println(tlog.Log()+"response status code :", strconv.Itoa(response.StatusCode))
			continue
		}
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf(tlog.Log(), "%s", err)
		}
		// jq like library to get the value of a jason field like jq -r ".messages"
		op, err := jq.Parse(".messages")
		if err != nil {
			log.Printf(tlog.Log(), "%s", err)
		}
		messages, err := op.Apply(contents)
		if err != nil {
			log.Printf(tlog.Log(), "%s", err)
		}
		fmt.Println(tlog.Log() + "Total messages in " + q + ":" + string(messages))
		qsize, _ := strconv.Atoi(string(messages))
		threshold, _ := strconv.Atoi(qMaxSizeThresholdAllow)
		// Use this environment variable to let run the service the first time to try to clear queues in case
		// it was restarted due to un process messages in the queues
		secondProbe := os.Getenv("SECOND_PROBE_FLAG")
		if secondProbe == "" {
			os.Setenv("SECOND_PROBE_FLAG", "true")
			return nil
		}
		if qsize > threshold {
			return errors.New(q + " messages(" + string(messages) + ") exceeds the threshold of " + qMaxSizeThresholdAllow)
		}
	}
	log.Print(tlog.Log() + "Successfully completed probe for rabbitMQ queue sizes")
	return nil
}

// PingHost will ping PG database endpoint set as secret in the evironment.yaml file such as development.yaml
// if successful ping continue otherwise will return an error
func PingHost(pgHost string, pgPort string) error {
	log.Println(tlog.Log()+"Starting liveness probe for PG host:", pgHost+":"+pgPort)
	seconds := 5
	timeout := time.Duration(seconds) * time.Second
	_, err := net.DialTimeout("tcp", pgHost+":"+pgPort, timeout)

	if err != nil {
		log.Print(tlog.Log()+"Failed to ping host:", pgHost+":"+pgPort, " error:", err.Error())
		return err
	}
	log.Print(tlog.Log()+"Successfully pinged host: ", pgHost+":"+pgPort)
	return nil
}

// GetRMQCert get the CA certificate from the TSL MQ rabbit value store in Vault RABBITMQ_TLS_CERT
// tlsCert store as base 64
func GetRMQCert(tlsCert string) (*x509.CertPool, error) {
	var caCertPool *x509.CertPool
	var decodedCA []byte
	//Certificate is stored in base64 in Vault
	decodedCA, err := base64.StdEncoding.DecodeString(tlsCert)
	if err != nil {
		return nil, err
	}
	caCertPool = x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM([]byte(decodedCA))
	if !ok {
		return nil, err
	}
	return caCertPool, nil
}
