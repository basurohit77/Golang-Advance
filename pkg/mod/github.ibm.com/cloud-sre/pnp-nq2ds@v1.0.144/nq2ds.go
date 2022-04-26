package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/ossmon"
	instana "github.com/instana/go-sensor"
	newrelic "github.com/newrelic/go-agent"
	"github.com/streadway/amqp"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-nq2ds/handlers"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	"github.ibm.com/cloud-sre/pnp-nq2ds/producer"
	"github.ibm.com/cloud-sre/pnp-nq2ds/shared"
	rabbitmq "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

var (
	messageProducer *producer.Producer
	chgStr          = regexp.MustCompile(`^CHG[0-9]+$`)
	nrApp           newrelic.Application
	sensor          *instana.Sensor
	mon             ossmon.OSSMon
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	var (
		environment             = os.Getenv("KUBE_APP_DEPLOYED_ENV")
		region                  = os.Getenv("KUBE_CLUSTER_REGION")
		url                     = os.Getenv("NQ_URL")
		url2                    = os.Getenv("NQ_URL2")
		rmqEnableMessages       = os.Getenv("RABBITMQ_ENABLE_MESSAGES")
		rmqAMQPSEndpoint        = os.Getenv("RABBITMQ_AMQPS_ENDPOINT")
		rmqTLSCert              = os.Getenv("RABBITMQ_TLS_CERT")
		qKey                    = os.Getenv("NQ_QKEY")
		exchangeName            = os.Getenv("NQ_EXCHANGE_NAME")
		caseOutQKey             = os.Getenv("NQ_CASE_OUT_QKEY")
		incidentOutQKey         = os.Getenv("NQ_INCIDENT_OUT_QKEY")
		incidentBulkOutQKey     = os.Getenv("NQ_INCIDENT_BULK_OUT_QKEY")
		maintenanceOutQKey      = os.Getenv("NQ_MAINTENANCE_OUT_QKEY")
		resourceOutQKey         = os.Getenv("NQ_RESOURCE_OUT_QKEY")
		notificationOutKey      = os.Getenv("NQ_NOTIFICATION_OUT_QKEY")
		notificationToSubOutKey = os.Getenv("NQ_NOTIFICATION_OUT_SUB_KEY")
		monitoringKey           = os.Getenv("NR_LICENSE")
		monitoringAppName       = os.Getenv("NR_APPNAME")
		instanaServiceName      = os.Getenv("INSTANA_SERVICE_NAME")
	)

	log.Print(tlog.Log()+"Environment: ", environment)
	log.Print(tlog.Log()+"Region: ", region)
	log.Print(tlog.Log()+"Notification queue url: ", strings.Split(url, "@")[1])
	log.Print(tlog.Log()+"Notification queue using Messages for RabbitMQ: ", isTargetMessagesForRabbitMQ(rmqEnableMessages))
	if rmqAMQPSEndpoint != "" {
		log.Print(tlog.Log()+"Notification queue amqps endpoint: ", strings.Split(rmqAMQPSEndpoint, "@")[1])
	} else {
		log.Print(tlog.Log()+"Notification queue amqps endpoint: ", rmqAMQPSEndpoint)
	}
	log.Print(tlog.Log()+"Notification queue input qkey: ", qKey)
	log.Print(tlog.Log()+"Notification queue exchange name: ", exchangeName)
	log.Print(tlog.Log()+"Notification queue Case output qkey: ", caseOutQKey)
	log.Print(tlog.Log()+"Notification queue Incident output qkey: ", incidentOutQKey)
	log.Print(tlog.Log()+"Notification queue Incident bulk output qkey: ", incidentBulkOutQKey)
	log.Print(tlog.Log()+"Notification queue Maintenance output qkey: ", maintenanceOutQKey)
	log.Print(tlog.Log()+"Notification queue Notification output qkey: ", notificationOutKey)
	log.Print(tlog.Log()+"Notification queue Notification to subscription output qkey: ", notificationToSubOutKey)
	log.Print(tlog.Log()+"Notification queue Resource output qkey: ", resourceOutQKey)
	log.Println(tlog.Log() + "Setting up connectivity to the database")
	err := shared.ConnectDatabase()

	if err != nil {
		log.Fatalln(tlog.Log() + err.Error())
	}

	if err := db.IsActive(shared.DBConn); err != nil {
		log.Fatalln(err.Error())
	}

	defer db.Disconnect(shared.DBConn)
	nrConfig := newrelic.NewConfig(monitoringAppName, monitoringKey)
	nrApp, err = newrelic.NewApplication(nrConfig)
	if err != nil {
		log.Print(tlog.Log() + err.Error())
	}

	// Initialize Instana sensor
	sensor = instana.NewSensor(instanaServiceName)
	mon.NewRelicApp = nrApp
	mon.Sensor = sensor

	// goroutine to test connectivity to database, and set up new connection when it is unhealthy
	go checkDBConnectivity(&mon)
	// initialize notifications adapter
	handlers.InitNotificationsAdapter()
	log.Print("Creating producer ...")
	var urls []string
	if isTargetMessagesForRabbitMQ(rmqEnableMessages) {
		urls = append(urls, rmqAMQPSEndpoint)
	} else {
		if url != "" {
			urls = append(urls, url)
		}
		if url2 != "" {
			urls = append(urls, url2)
		}
	}
	// set rabbitmq details to post notification
	handlers.RabbitmqURLs = urls
	handlers.RabbitmqTLSCert = rmqTLSCert
	handlers.RabbitmqEnableMessages = isTargetMessagesForRabbitMQ(rmqEnableMessages)
	handlers.NotificationRoutingKey = notificationOutKey
	handlers.MQExchangeName = exchangeName

	if isTargetMessagesForRabbitMQ(rmqEnableMessages) {
		messageProducer, err = producer.NewSSLProducer(urls, rmqTLSCert, exchangeName, "direct", caseOutQKey, &mon)
	} else {
		messageProducer, err = producer.NewProducer(urls, exchangeName, "direct", caseOutQKey, &mon)
	}
	if err != nil {
		log.Print(tlog.Log(), err)
	} else {
		messageProducer.CaseRoutingKey = caseOutQKey
		messageProducer.IncidentRoutingKey = incidentOutQKey
		messageProducer.IncidentBulkRoutingKey = incidentBulkOutQKey
		messageProducer.MaintenanceRoutingKey = maintenanceOutQKey
		messageProducer.ResourceRoutingKey = resourceOutQKey
		messageProducer.NotificationRoutingKey = notificationOutKey
		messageProducer.NotificationToSubRoutingKey = notificationToSubOutKey
	}

	log.Print(tlog.Log() + "Connecting to message queue ...")
	var c *rabbitmq.AMQPConsumer
	if isTargetMessagesForRabbitMQ(rmqEnableMessages) {
		c = rabbitmq.NewSSLConsumer(urls, rmqTLSCert, qKey, exchangeName)
	} else {
		c = rabbitmq.NewConsumer(urls, qKey, exchangeName)
	}

	cName := "nq2ds_" + monitor.NRenvironment + "_" + monitor.NRregion
	c.Name = cName
	// server automatically acks a msg that is consumed from a queue
	// no need to call msg.Ack(false) if AutoAck is true
	c.AutoAck = false
	log.Print(tlog.Log() + "Listening for messages on queues ...")
	go c.Consume(f)
	// healthz for Liveness probe using on oss-charts
	http.HandleFunc("/healthz", shared.LivenessProbe)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func f(msg amqp.Delivery) {

	if msg.RoutingKey == "incident" {
		log.Print(tlog.Log() + "Consumed new or updated incident from queue")
		incident, isBulkLoad, isBadMessage := handlers.ProcessIncident(shared.DBConn, msg.Body, &mon)
		if isBadMessage {
			log.Print(tlog.Log() + "Bad message. Sending to dead letter queue")
			if err := msg.Reject(false); err != nil {
				log.Println(tlog.Log(), err)
			}
		} else {
			postErr := messageProducer.PostIncident(incident, isBulkLoad, &mon)
			if postErr != nil {
				log.Print(tlog.Log()+"Error posting incident message. Sending request message to dead letter queue. Err = ", postErr)
				err := msg.Reject(false)
				if err != nil {
					log.Println(tlog.Log(), err)
				}
			} else {
				ackErr := msg.Ack(false)
				if ackErr != nil {
					log.Print(tlog.Log() + "Error acking incident message: " + ackErr.Error())
				}
			}
		}
	} else if msg.RoutingKey == "maintenance" {
		log.Print(tlog.Log() + "Consumed new or updated maintenance from queue")
		messageMap, decryptedMessage := handlers.DecodeAndMapMaintenceMessage(msg.Body, &mon)

		log.Println(tlog.Log(), api.RedactAttributes(messageMap))
		msgRejected := false
		posts := 0
		postErrs := 0
		notificationPosts := 0
		notificationPostErrs := 0
		isFromSN := false
		isBulk := false
		if _, ok := messageMap["result_from_sn"]; ok {
			isFromSN = true
			isBulk = true
		}
		if !isFromSN {
			// check if the number attribute exists. If so, assign accordingly
			if _, ok := messageMap["number"]; ok {
				isFromSN = true
			}
		}
		// check if source_id is a change request record of the form: ^CHG[0-9]+$
		if sourceID, ok := messageMap["source_id"]; ok {
			isChange := chgStr.MatchString(sourceID.(string))
			if isChange {
				isFromSN = true
			}
		}
		if isFromSN {
			maintenanceMaps, notifications := handlers.ProcessSNMaintenance(shared.DBConn, decryptedMessage, isBulk, &mon)
			for i := range maintenanceMaps {
				postErr := messageProducer.PostMaintenance(&maintenanceMaps[i], &mon)
				if postErr != nil {
					postErrs++
					log.Print(tlog.Log()+"Error posting maintenance message. Sending request message to dead letter queue. Err = ", postErr, maintenanceMaps[i].RecordID, maintenanceMaps[i].SourceID)
					if !msgRejected {
						err := msg.Reject(false)
						if err != nil {
							log.Println(tlog.Log(), err)
						}
						msgRejected = true
					}
				} else {
					posts++
				}
			}
			for i := range notifications {
				postErr := messageProducer.PostNotification(&notifications[i], &mon) //Fixed G601 (CWE-118): Implicit memory aliasing in for loop: postErr := messageProducer.PostNotification(&v)
				if postErr != nil {
					notificationPostErrs++
					log.Print(tlog.Log()+"Error posting notification message. Sending request message to dead letter queue. Err = ", postErr, notifications[i].IncidentID, notifications[i].SourceID)
					if !msgRejected {
						errReject := msg.Reject(false)
						if errReject != nil {
							log.Println(tlog.Log()+"Error rejecting message: ", errReject)
						}
						msgRejected = true
					}
				} else {
					notificationPosts++
				}
			}
			if !msgRejected {
				ackErr := msg.Ack(false)
				if ackErr != nil {
					log.Print(tlog.Log() + "Error acking status message: " + ackErr.Error())
				}
			}
			log.Printf(tlog.Log()+"Total maintenance posts: %d \n\tTotal in error: %d ", posts, postErrs)
			log.Printf(tlog.Log()+"Total notification posts: %d \n\tTotal in error: %d ", notificationPosts, notificationPostErrs)
			return
		}
		ackErr := msg.Ack(false)
		if ackErr != nil {
			log.Print(tlog.Log() + "Error acking maintenance message: " + ackErr.Error())
		}
	} else if msg.RoutingKey == "resource" {
		msgSnippet := ""
		if len(string(msg.Body)) > 40 {
			msgSnippet = string(msg.Body)[0:40]
		} else {
			msgSnippet = string(msg.Body)
		}
		log.Println(tlog.Log()+"Consumed new or updated resources from queue :", msgSnippet+"...")
		updateResources, err := handlers.ProcessResourceMsg(shared.DBConn, msg.Body, &mon)
		if err == nil {
			if len(updateResources) > 0 {
				log.Println(tlog.Log()+"updated resources :", len(updateResources), updateResources[0].SourceID)

			} else {
				log.Println(tlog.Log()+"updated resources :", len(updateResources))
			}
			postErr := messageProducer.PostResources(updateResources, &mon)
			if postErr != nil {
				log.Print(tlog.Log()+"Error posting resource message. Sending request message to dead letter queue. Err = ", postErr)
				errReject := msg.Reject(false)
				if errReject != nil {
					log.Println(tlog.Log()+"Error rejecting message: ", errReject)
				}
			} else {
				ackErr := msg.Ack(false)
				if ackErr != nil {
					log.Print(tlog.Log() + "Error acking resource message: " + ackErr.Error())
				}
			}
		} else {
			log.Print(tlog.Log()+"Error posting resource message. Sending request message to dead letter queue. Err = ", err)
			errReject := msg.Reject(false)
			if errReject != nil {
				log.Println(tlog.Log()+"Error rejecting message: ", errReject)
			}
		}
	} else if msg.RoutingKey == "status" {
		log.Print(tlog.Log() + "Consumed new or updated status from queue")
		isBadMessage := handlers.ProcessStatus(shared.DBConn, msg.Body, &mon)
		if isBadMessage {
			log.Print(tlog.Log() + "Bad message. Sending to dead letter queue")
			errReject := msg.Reject(false)
			if errReject != nil {
				log.Println(tlog.Log()+"Error rejecting message: ", errReject)
			}
		} else {
			ackErr := msg.Ack(false)
			if ackErr != nil {
				log.Print(tlog.Log() + "Error acking status message: " + ackErr.Error())
			}
		}
	} else if msg.RoutingKey == "notification" {
		log.Print("Consumed new or updated notification from queue")
		notification, err := handlers.ProcessNotification(shared.DBConn, msg.Body, &mon)
		if err != nil {
			log.Print(tlog.Log() + "Bad notification message. Sending to dead letter queue")
			if err := msg.Reject(false); err != nil {
				log.Println(tlog.Log(), err)
			}
		} else {
			postErr := messageProducer.PostNotificationToSubscriptionConsumer(notification, &mon)
			if postErr != nil {
				// Should never happen, because the post above will retry indefinitely
				log.Print(tlog.Log()+"Error posting notification message to subscription consumer. Sending request message to dead letter queue. Err = ", postErr)
				if err := msg.Reject(false); err != nil {
					log.Println(tlog.Log(), err)
				}
			} else {
				// All is good, acknowledge the message
				ackErr := msg.Ack(false)
				if ackErr != nil {
					log.Print(tlog.Log() + "Error acking notification message: " + ackErr.Error())
				}
			}
		}
	}
}

func isTargetMessagesForRabbitMQ(rmqEnableMessages string) bool {
	return rmqEnableMessages == "true"
}

func checkDBConnectivity(mon *ossmon.OSSMon) {
	for {
		span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
		txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRdbConnection, nil, nil)
		ctx = newrelic.NewContext(ctx, txn)
		if err := db.IsActive(shared.DBConn); err != nil {
			log.Println(tlog.Log(), err.Error())
			ossmon.SetTagsKV(ctx, ossmon.TagEnv, monitor.ENVIRONMENT,
				ossmon.TagRegion, monitor.REGION,
				ossmon.TagZone, monitor.ZONE,
				"dbError", err.Error(),
				handlers.DBFailedErr, true,
				"apiKubeClusterRegion", monitor.NRregion,
				"apiKubeAppDeployedEnv", monitor.NRenvironment)
			ossmon.SetError(ctx, handlers.DBFailedErr+"-"+err.Error())
			db.Disconnect(shared.DBConn)
			log.Println(tlog.Log() + "Reconnecting to the database")
			err = shared.ConnectDatabase()
			if err != nil {
				log.Println(tlog.Log(), err.Error())
			}
		}
		ossmon.SetTagsKV(ctx, ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagZone, monitor.ZONE,
			"dbError", "No errors",
			handlers.DBFailedErr, false,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment)
		span.Finish()
		err:=txn.End()
		if err!=nil {
			log.Println(tlog.Log(),err)
		}
		time.Sleep(10 * time.Second)
	}
}


