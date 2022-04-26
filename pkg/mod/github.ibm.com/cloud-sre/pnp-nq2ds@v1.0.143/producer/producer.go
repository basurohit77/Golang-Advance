package producer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	"github.ibm.com/cloud-sre/pnp-nq2ds/handlers"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	rabbitmq "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

// Producer - Handles writting to queues.
type Producer struct {
	url                         []string
	ExchangeName                string
	ExchangeType                string
	InitialRoutingKey           string
	CaseRoutingKey              string
	IncidentRoutingKey          string
	IncidentBulkRoutingKey      string
	MaintenanceRoutingKey       string
	ResourceRoutingKey          string
	NotificationRoutingKey      string
	NotificationToSubRoutingKey string

	producer *rabbitmq.AMQPProducer
}

// NewProducer - constructor for producer
func NewProducer(url []string, exchangeName string, exchangeType string, initialRoutingKey string,mon *ossmon.OSSMon) (*Producer, error) {
	return internalNewProducer(url, "", exchangeName, exchangeType, initialRoutingKey,mon)
}

// NewSSLProducer - constructor for producer that uses SSL
func NewSSLProducer(url []string, tlsCert string, exchangeName string, exchangeType string, initialRoutingKey string,mon *ossmon.OSSMon) (*Producer, error) {
	return internalNewProducer(url, tlsCert, exchangeName, exchangeType, initialRoutingKey,mon)
}

// internalNewProducer - constructor for producer
func internalNewProducer(url []string, tlsCert string, exchangeName string, exchangeType string, initialRoutingKey string,mon *ossmon.OSSMon) (*Producer, error) {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()

	producer := &Producer{
		url:               url,
		ExchangeName:      exchangeName,
		ExchangeType:      exchangeType,
		InitialRoutingKey: initialRoutingKey,
	}
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
	)

	if tlsCert != "" {
		producer.producer = rabbitmq.NewSSLProducer(producer.url, tlsCert, producer.InitialRoutingKey, producer.ExchangeName, producer.ExchangeType)
	} else {
		producer.producer = rabbitmq.NewProducer(producer.url, producer.InitialRoutingKey, producer.ExchangeName, producer.ExchangeType)
	}
	err := producer.producer.Connect()
	if err != nil {
		ossmon.SetError(ctx, tlog.FuncName()+"-"+err.Error()) //Re
		return producer, err
	}
	return producer, nil
}

// PostIncident - sends a Incident message to the appropriate routing key
func (producer Producer) PostIncident(incidentReturn *datastore.IncidentReturn, isBulkLoad bool, mon *ossmon.OSSMon) error {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())

	defer span.Finish()
	// Record start time:
	startTime := time.Now()
	if incidentReturn == nil {
		log.Print(tlog.Log() + "Post not needed since incident is nil")
		ossmon.SetTagsKV(ctx,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
			"pnp-StartTime", startTime.Unix(),
			"pnp-Kind", "incident",
			"pnp-Operation", monitor.SrvPrfx+" post not needed since incident is nil",
		)
		return nil
	}
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRpostIncident, nil, nil)
	ctx = newrelic.NewContext(ctx, txn)
	defer func() {
		err:= txn.End()
		if err !=nil {
			log.Println(tlog.Log(),err)
		}
	}()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
		"pnp-Source", incidentReturn.Source,
		"pnp-SourceID", incidentReturn.SourceID,
		"pnp-Kind", "incident",
		"pnp-Operation", monitor.SrvPrfx+"postmsg",
	)
	log.Print(tlog.Log()+"Going to post: ", api.RedactAttributes(incidentReturn))
	// Struct to JSON string:
	bytes, err := json.Marshal(incidentReturn)
	if err != nil {
		// Should never happen
		return err
	}
	messageToPost := string(bytes)
	// Post message:
	if isBulkLoad {
		producer.postMessageWithRetry(ctx, producer.IncidentBulkRoutingKey, messageToPost)
	} else {
		producer.postMessageWithRetry(ctx, producer.IncidentRoutingKey, messageToPost)
	}
	// Determine duration to report to NewRelic and Instana
	ossmon.SetTag(ctx, "pnp-Duration", time.Since(startTime))
	return nil
}

// PostMaintenance - sends a Maintenance message to the appropriate routing key
func (producer Producer) PostMaintenance(maintenanceMap *datastore.MaintenanceMap, mon *ossmon.OSSMon) error {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	// Record start time:
	startTime := time.Now()
	if maintenanceMap == nil {
		log.Print(tlog.Log() + "Post not needed since maintenance is nil")
		ossmon.SetTagsKV(ctx,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
			"pnp-StartTime", startTime.Unix(),
			"pnp-Kind", "maintenance",
			"pnp-Operation", monitor.SrvPrfx+"post not needed since maintenance is nil",
		)
		return nil
	}
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRpostMaintenance, nil, nil)
	ctx = newrelic.NewContext(ctx, txn)
	defer func() {
		err:= txn.End()
		if err !=nil {
			log.Println(tlog.Log(),err)
		}
	}()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
		"pnp-Source", maintenanceMap.Source,
		"pnp-SourceID", maintenanceMap.SourceID,
		"pnp-Kind", "maintenance",
		"pnp-Operation", monitor.SrvPrfx+"postmsg",
	)
	log.Print(tlog.Log()+"Going to post: ", api.RedactAttributes(maintenanceMap))
	// Struct to JSON string:
	bytes, err := json.Marshal(maintenanceMap)
	if err != nil {
		// Should never happen
		return err
	}
	messageToPost := string(bytes)
	// Post message:
	producer.postMessageWithRetry(ctx, producer.MaintenanceRoutingKey, messageToPost)
	// Determine duration to report to NewRelic and Instana
	ossmon.SetTag(ctx, "pnp-Duration", time.Since(startTime))
	return nil
}

// PostNotification - sends a Notification message to the appropriate routing key
func (producer Producer) PostNotification(notification *datastore.NotificationMsg, mon *ossmon.OSSMon) error {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	// Record start time:
	startTime := time.Now()
	if notification == nil {
		log.Print(tlog.Log() + "Post not needed since notification is nil")
		ossmon.SetTagsKV(ctx,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
			"pnp-StartTime", startTime.Unix(),
			"pnp-Kind", "notification",
			"pnp-Operation", monitor.SrvPrfx+"post not needed since notification is nil",
		)
		return nil
	}
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRpostNotification, nil, nil)
	ctx = newrelic.NewContext(ctx, txn)
	defer func() {
		err:= txn.End()
		if err !=nil {
			log.Println(tlog.Log(),err)
		}
	}()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
		"pnp-Source", notification.Source,
		"pnp-SourceID", notification.SourceID,
		"pnp-Kind", "notification",
		"pnp-Operation", monitor.SrvPrfx+"postmsg",
	)
	log.Print(tlog.Log()+"DEBUG: Going to post: ", notification)
	// Struct to JSON string:
	bytes, err := json.Marshal(notification)
	if err != nil {
		// Should never happen
		return err
	}
	messageToPost := string(bytes)
	// Post message:
	producer.postMessageWithRetry(ctx, producer.NotificationRoutingKey, messageToPost)
	// Determine duration to report to NewRelic and Instana
	ossmon.SetTag(ctx, "pnp-Duration", time.Since(startTime))
	return nil
}

// PostNotificationToSubscriptionConsumer - sends a Notification message to the subscription consumer
func (producer Producer) PostNotificationToSubscriptionConsumer(notification *datastore.NotificationMsg, mon *ossmon.OSSMon) error {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	// Record start time:
	startTime := time.Now()
	if notification == nil {
		log.Print(tlog.Log() + "Post to subscription consumer is not needed since notification is nil")
		ossmon.SetTagsKV(ctx,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
			"pnp-StartTime", startTime.Unix(),
			"pnp-Kind", "notification",
			"pnp-Operation", monitor.SrvPrfx+"post to subscription consumer is not needed since notification is nil",
		)
		return nil
	}
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRpostNotifSubConsumer, nil, nil)
	ctx = newrelic.NewContext(ctx, txn)
	defer func() {
		err:= txn.End()
		if err !=nil {
			log.Println(tlog.Log(),err)
		}
	}()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
		"pnp-Source", notification.Source,
		"pnp-SourceID", notification.SourceID,
		"pnp-Kind", "notification",
		"pnp-Operation", monitor.SrvPrfx+"postmsg",
	)
	log.Print(tlog.Log()+"DEBUG: Going to post: ", notification)
	// Struct to JSON string:
	bytes, err := json.Marshal(notification)
	if err != nil {
		// Should never happen
		return err
	}
	messageToPost := string(bytes)
	// Post message:
	producer.postMessageWithRetry(ctx, producer.NotificationToSubRoutingKey, messageToPost)
	// Determine duration to report to NewRelic and Instana
	ossmon.SetTag(ctx, "pnp-Duration", time.Since(startTime))
	return nil
}

// PostResources - sends one or more Resource message to the appropriate routing key
func (producer Producer) PostResources(resources []*datastore.ResourceReturn, mon *ossmon.OSSMon) error {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	var errToReturn error
	// Record start time:
	startTime := time.Now()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
		"pnp-Kind", "resource",
		"pnp-Operation", monitor.SrvPrfx+"postmsg",
	)
	// Try all the resources and if there is one or more errors, capture one of the errors to return:
	if resources != nil {
		for _, resource := range resources {
			err := producer.PostResource(resource, mon)
			if err != nil {
				errToReturn = err
			}
		}
	}
	return errToReturn
}

// PostResource - sends a Resource message to the appropriate routing key
func (producer Producer) PostResource(resourceReturn *datastore.ResourceReturn, mon *ossmon.OSSMon) error {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	// Record start time:
	startTime := time.Now()
	if resourceReturn == nil {
		log.Print(tlog.Log() + "Post not needed since resource is nil")
		ossmon.SetTagsKV(ctx,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
			"pnp-StartTime", startTime.Unix(),
			"pnp-Kind", "resource",
			"pnp-Operation", monitor.SrvPrfx+"post not needed since resource is nil",
		)
		return nil
	}
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRpostResource, nil, nil)
	ctx = newrelic.NewContext(ctx, txn)
	defer func() {
		err:= txn.End()
		if err !=nil {
			log.Println(tlog.Log(),err)
		}
	}()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
		"pnp-Source", resourceReturn.Source,
		"pnp-SourceID", resourceReturn.SourceID,
		"pnp-Kind", "resource",
		"pnp-Operation", monitor.SrvPrfx+"postmsg",
	)
	log.Print(tlog.Log()+"DEBUG: Going to post: ", resourceReturn)
	// Struct to JSON string:
	bytes, err := json.Marshal(resourceReturn)
	if err != nil {
		// Should never happen
		return err
	}
	messageToPost := string(bytes)
	// Post message:
	producer.postMessageWithRetry(ctx, producer.ResourceRoutingKey, messageToPost)
	// Determine duration to report to monitor
	ossmon.SetTag(ctx, "pnp-Duration", time.Since(startTime))
	return nil
}

func (producer Producer) postMessageWithRetry(ctxParent context.Context, routingKey string, messageToPost string) {
	err := errors.New("") // dummy error just to ensure loop below executes at least once
	for err != nil {
		err = producer.postMessage(ctxParent, routingKey, messageToPost)
		if err != nil {
			log.Print(tlog.Log()+"sleep and retry, error: ", err)
			time.Sleep(time.Second * 5)
		}
	}
}

func (producer Producer) postMessage(ctxParent context.Context, routingKey string, messageToPost string) error {
	log.Print(tlog.Log()+"Posting the following message to "+routingKey+": ", messageToPost)
	producer.producer.RoutingKey = routingKey
	// Encrypt message to post:
	encryptedData, err := encryption.Encrypt(messageToPost)
	if err != nil {
		// Should never happen
		log.Print(tlog.Log()+"Error occurred trying to encrypt data, err = ", err)
		ossmon.SetTag(ctxParent, handlers.EncryptionErr, err.Error())
		ossmon.SetError(ctxParent, handlers.EncryptionErr+"-"+err.Error()) //Replaces NewRelic NQRL monitor
		return err
	}
	return producer.producer.Produce(string(encryptedData))
}
