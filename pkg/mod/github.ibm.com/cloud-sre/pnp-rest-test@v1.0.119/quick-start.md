# Quick Start

This document provides the quick start steps to launch pnp-rest-test locally.

## Step 1 - Install declarative deployment tools
Make sure declarative deployment tools is downloaded to `$GOPATH/src/github.ibm.com/cloud-sre/declarative-deployment-tools`

## Step 2 - Install postgres
Make sure you have Postgres install on your testing laptop

## Step 3 - Update envs.sh
Edit `envs.sh` to update any change on $SERVER_KEY or Postgres password.

```
$ vim ./envs.sh
$ source ./envs.sh
```

## Step 4 - Build the pnp-rest-test
```
$ gomake deps
$ gomake
```

## Step 5 - Run tests

### Option 1 - Run specific tests

```
$ ./main --test incident --token $SERVER_KEY
$ ./main --test incident --token $SERVER_KEY --endpoint https://pnp-api-oss.test.cloud.ibm.com/catalog/api/info
```

### Option 2 - Run the testing server

```
$ ./main -startServer=true
```

Then you can try to use following commands to trigger tests:

```
$ curl http://localhost:8000/run/\?basic
$ curl http://localhost:8000/run/\?RunSubscriptionFull
```

Some other commands:

```
$ curl http://localhost:8000
$ curl -X POST http://localhost:8000 -d '{"foo": "bar"}'
$ curl http://localhost:8000/help
```