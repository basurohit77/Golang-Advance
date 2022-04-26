[![Build Status](https://wcp-cto-sre-jenkins.swg-devops.com/buildStatus/icon?job=Pipeline/api-pnp-rabbitmq-connector/master)](https://wcp-cto-sre-jenkins.swg-devops.com/job/Pipeline/job/api-pnp-rabbitmq-connector/job/master/)

# pnp-rabbitmq-connector

This is a package that helps setting up a consumer and a producer clients for a RabbitMQ server.  

This package uses the protocol AMQP(Advanced Message Queuing Protocol) for communication with the RabbitMQ server.  


## Unit tests, coverage and scanning

- Run `make test` to test.
- To view unit test coverage run `go tool cover -html=coverage.out` after running `make test`.
- You should aim to get at least 80% coverage for each package.
- Run `make scan` to run a security scan.
- Unit tests and scan should be successful before submitting a pull request to the master branch.
- You can find results of unit tests in unittest.out.

## CI/CD

- The jenkins job will run unit tests and a security scan on pull requests and pulls.
- Upon successful merge to master, the list of jobs in config-jenkins.yaml will be triggered to rebuild the dependant components.


## Using this lib

See the [examples](./_examples) directory for client implementation of Producer and Consumer.  

Once you've downloaded this package and imported it into your code, you can create a consumer and producer as follows.

### Consumer

The consumer automatically creates the queues and bind each queue to the specified exchange if they do not exist.  
If the exchange does not exist, the consumer will exit with an error.  
You need the following info in order to set up a consumer:

1. RabbitMQ URL with username and password, you can spcify up to 2 URL's  
    `var url = []string{"amqp://guest:guest@192.168.99.100:5672"}`
2. Queue(s) and Routing Key(s)  
    You can specify one or more queues for a consumer to read messages from
    // qKey holds the list of all queues and routing keys that msgs should be consumed from
    // the string is separated by comma(","), between each comma there a pair of queue:key value
    // e.g. "incident.subscription:incident", the queue name is "incident.subscription" and the routing key is "incident"  
    `var qKey = "incident.subscription:incident,incident.status:incident,incident.nq2ds:incident"`
3. Exchange Name  
    The exchange name is needed to bind a queue to an Exchange  
    `var exchangeName = "pnp.direct"`

### Producer

The producer automatically creates the exchange if it does not exists.  
You need the following info in order to set up a producer:

1. RabbitMQ URL with username and password, you can spcify up to 2 URL's  
    `var url          = []string{"amqp://guest:guest@192.168.99.100:5672"}`
2. Routing Key  
    `var routingKey   = "incident"`
3. Exchange Name  
    `exchangeName = "pnp.direct"`
4. Exchange Type  
    `var exchangeType = "direct"`


### Encryption/Decryption

The consumers in the [examples](./_examples) directory use decryption when they receive a messasge.  
That means the msg must have been encrypted the a producer.  
Both producer and consumer must have the same master key.  
You need to set the environment variable `MASTER_KEY` before starting the process.  

e.g. `export MASTER_KEY=000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000`  

Note: the master key must be in hexadecimal format
