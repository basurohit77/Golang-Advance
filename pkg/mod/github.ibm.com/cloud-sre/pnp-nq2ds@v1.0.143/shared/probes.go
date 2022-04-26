package shared

import (
	"crypto/tls"
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

var (
	qSizeThreshold   = os.Getenv("RABBITMQ_THRESHOLD")
	rabbitAMEndPoint = os.Getenv("RABBITMQ_AMQPS_ENDPOINT")
	nqQKey           = os.Getenv("NQ_QKEY")
	rMQHost          = os.Getenv("RABBITMQ_HOST")
	rmqTLSCert       = os.Getenv("RABBITMQ_TLS_CERT")
	nq2dsQueues      = []string{"incident.nq2ds", "maintenance.nq2ds", "status.nq2ds", "resource.nq2ds", "notification.nq2ds"}
)

//LivenessProbe is use as kubernetes probe to check the health of the pod the liveness defintion is located at chart under templates/deployment file and will look like
// livenessProbe:
//             httpGet:
//               path: /healthz
//               port: {{ .Values.nq2dsPort }}
//               httpHeaders:
//                - name: Custom-Header
//                  value: Alive
//             initialDelaySeconds: 10
//             periodSeconds: 60
// for more information check https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
// This probe will check three critical point
// PostgreSQL connectivity, it is alive? able to reach/connect?
// rabbitMQ connectivity, it is alive? able to reach/connect?
// rabbitQM.messages will check if the number of messages is lower than the threshold set
// It Any of the above fails Kubernetes will try up to three times (as default) otherwise will restart the pod
// Liveness won't check the queue sizes in the first pass to let the application to cleanup any pending messages
func LivenessProbe(w http.ResponseWriter, r *http.Request) {
	log.Print(tlog.Log() + "Starting liveness probes")
	//Cerificate is stored in base64 in Vault
	decoded_ca, err := base64.StdEncoding.DecodeString(rmqTLSCert)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", err.Error())))
		log.Print(tlog.Log()+"liveness probe decode certification failed, error:", err.Error())
	}
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM([]byte(decoded_ca))
	if !ok {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", err.Error())))
		log.Print(tlog.Log()+"liveness probe failed to parse encoded certificate, error:", err.Error())
	}
	tlsConfig := &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	c := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	if err := pingHost(); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", err.Error())))
		log.Print(tlog.Log()+"liveness probe failed to ping PG host, error:", err.Error())
	} else if err := rabbitMQHealthz(c); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", err.Error())))
		log.Print(tlog.Log()+"liveness probe failed to ping rabbitMQ host, error:", err.Error())
	} else if err := rabbitMQSize(c); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", err.Error())))
		log.Print(tlog.Log() + "liveness probe, error:" + err.Error())
	} else {
		log.Print(tlog.Log() + "Liveness probes successfully completed")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}

}

//getQName Parse the queue names from the enviroment variable named nqQKey set at the values.yaml
// nqQKey has the format as incident.nq2ds:incident where queue name 'incident.nq2ds' and :incident is the object need to remove the object part
func getQName() []string {
	var qNames []string
	//if nqQKey is not set will use the local variable nq2dsQueues
	if nqQKey == "" {
		return nq2dsQueues
	}
	qNames = strings.Split(nqQKey, ",")
	for i, q := range qNames {
		qNames[i] = q[:strings.Index(q, ":")]
	}
	return qNames

}

//parseRMQEndPoint uses the rabbit enpoint set at each region as secret RABBITMQ_AMQPS_ENDPOINT the endpoint is formated as
// transport protocoluser:pwd@host like amqps://ibmxxxx:pwd12312@0023:port
func parseRMQEndPoint(ampEndPoint string) (user string, pwd string, host string) {
	ampEndPoint = ampEndPoint[len("amqps://"):] //removes the transport protocol
	user = ampEndPoint[0:strings.Index(ampEndPoint, ":")]
	strAux := ampEndPoint[strings.Index(ampEndPoint, ":")+1:]
	pwd = strAux[0:strings.Index(strAux, "@")]
	host = strAux[strings.Index(strAux, "@")+1:]
	return user, pwd, host
}

//rabbitMQHealthz does an http.Get to a rabbit host set at the region_deployment.yaml file as RABBITMQ_HOST
//if status is 200 then it is alive otherwise will return an error
func rabbitMQHealthz(c *http.Client) error {
	log.Println(tlog.Log()+"Starting probe for rabbitMQ host:", rMQHost)

	if rMQHost == "" {
		return errors.New("Invalid host name:" + rMQHost)
	}
	resp, err := c.Get(rMQHost)
	if err != nil {
		return errors.New(tlog.Log() + "Failed to ping rabbitMQ host:" + rMQHost + " error:" + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(tlog.Log() + "expecting status code of 200 got " + strconv.Itoa(resp.StatusCode) + " error:" + err.Error())
	}
	log.Println(tlog.Log()+"Successfully pinged rabbitMQ host: ", rMQHost)
	return nil
}

// rabbitMQSize check if the messages in the queue are less than the threshold set at the values.yaml file as RABBITMQ_THRESHOLD is any of messges the listed queues
// is grather than the threshold in will set an erroneos liveness to ask the pod to restart
func rabbitMQSize(client *http.Client) error {
	log.Println(tlog.Log() + "Starting probe for rabbitMQ queue sizes")

	user, pwd, _ := parseRMQEndPoint(rabbitAMEndPoint)
	for _, q := range getQName() {
		if strings.Contains(q, "case") {
			//cases are not longer in used
			continue
		}
		url := rMQHost + "api/queues/%2F/" + q
		if rMQHost[len(rMQHost)-1:] != "/" {
			url = rMQHost + "/api/queues/%2F/" + q
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
		threshold, _ := strconv.Atoi(qSizeThreshold)
		// Use this environment variable to let run the service the first time to try to clear queues in case
		// it was restarted due to un process messages in the queues
		secondProbe := os.Getenv("SECOND_PROBE_FLAG")
		if secondProbe == "" {
			os.Setenv("SECOND_PROBE_FLAG", "true")
			return nil
		}
		if qsize > threshold {
			return errors.New(q + " messges(" + string(messages) + ") excceds the threshold of " + qSizeThreshold)
		}
	}
	log.Print(tlog.Log() + "Successfully completed probe for rabbitMQ queue sizes")
	return nil
}

//pingHost will ping PG database endpoint set as secrent in the evironment.yaml file such as development.yaml
//if successful ping continue otherwise will return an error
func pingHost() error {
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
