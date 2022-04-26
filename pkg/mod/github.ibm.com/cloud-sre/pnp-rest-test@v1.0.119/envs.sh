#!/bin/bash

export GOPATH=$(go env GOPATH)
export DEPLOY_TOOL_PATH=$GOPATH/src/github.ibm.com/cloud-sre/declarative-deployment-tools
export PATH=$PATH:$GOPATH/bin:$DEPLOY_TOOL_PATH

export CATALOG_URL='https://pnp-api-oss.test.cloud.ibm.com'
export SERVER_KEY='<PNP-TOKEN>'

export DRApiKey=''
export HOOK_KEY='<Replace it with HOOK_KEY>'

export NQ_URL='amqps://runtime-user:<Replace-Me>@portal-ssl503-0.bmix-wdc-yp-f743dde7-8eab-45bd-97dc-da89723a28a5.3027868685.composedb.com:16348/bmix-wdc-yp-f743dde7-8eab-45bd-97dc-da89723a28a5' # pragma: whitelist secret
export NQ_URL2='amqps://runtime-user:<Replace-Me>@portal-ssl477-1.bmix-wdc-yp-f743dde7-8eab-45bd-97dc-da89723a28a5.3027868685.composedb.com:16348/bmix-wdc-yp-f743dde7-8eab-45bd-97dc-da89723a28a5' # pragma: whitelist secret

export PG_DB='pnptest'
export PG_DB_PASS='pnp' # pragma: whitelist secret
export PG_DB_USER='pnp' # pragma: whitelist secret
export PG_HOST='localhost'
export PG_DB_IP='127.0.0.1'
export PG_PORT='5432'
export PG_SSLMODE='disable'

export RMQHooksCase="$CATALOG_URL/pnphooks/api/v1/snow/cases"
export RMQHooksChange="$CATALOG_URL/pnphooks/api/v1/snow/changes"
export RMQHooksDr="$CATALOG_URL/pnphooks/api/v1/doctor/maintenances"
export RMQHooksIncident="$CATALOG_URL/pnphooks/api/v1/snow/incidents"
export subscriptionURL="$CATALOG_URL/pnpsubscription/api/v1/pnp/subscriptions/"

export SN_KEY='<Service-Now-Key>' # pragma: whitelist secret
export envType='dev'
export ingressIP=''
export snHost='watsondev'
export snToken="$SN_KEY"
