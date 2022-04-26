PRIVATE_MODULES=$(shell grep "^\s*github.ibm.com" go.mod | awk '{print $$1}')

cicd:
	git diff-index --quiet HEAD -- #error if uncommitted changes
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main
	$(eval imageRegistry = wcp-cto-oss-docker-local.artifactory.swg-devops.com/api-pnp-resource-adapter)
	$(eval imageTag = $(shell git rev-parse HEAD)$(shell date +%s))
	$(eval imageURI = "$(imageRegistry):$(imageTag)")
	@echo $(imageURI)
	docker build -t $(imageURI) .
	rm main
	docker push $(imageURI)

# run unit tests
.PHONY: test
test:
	@docker ps >> /dev/null -- #ersror if docker not running/not configured
	$(eval HOST = $(shell ipport=$${DOCKER_HOST#*//};  dbip=$${ipport%%:*}; if [[ $$dbip == "" ]]; then echo localhost; else echo $$dbip; fi))
	$(eval RABBITMQ_PORT = 5672)
	$(eval NQ_URL = amqp://guest:guest@$(HOST):5672) # pragma: whitelist secret
	$(eval routingKey = resource)
	$(eval exchangeName = pnp.direct)
	$(eval exchangeType = direct)
	$(eval qKey = resource.nq2ds:resource)
	$(eval snHost = watsondev)
	$(eval subscriptionURL = https://api-oss-dev.bluemix.net/pnpsubscription/api/v1/pnp/subscriptions/)


	@echo
	@echo "---> Building rabbitmq image"
	@docker build -f Dockerfile.rabbitmq -t pnp-rest-test-rabbitmq .
	@echo

	@echo
	@echo "---> Starting rabbitmq image"
	$(eval CID2 = $(shell docker run -d -p $(RABBITMQ_PORT):5672 --hostname localhost pnp-rest-test-rabbitmq))
	@echo "Started rabbitmq -  $(CID2)"

	@echo
	@echo "---> Installing dependencies"
	go get github.com/axw/gocov/...
	go get github.com/AlekSi/gocov-xml
	go get -u github.com/jstemmer/go-junit-report
	$(shell go env -w GOPRIVATE="*.ibm.com")
	$(foreach module,$(PRIVATE_MODULES), go get -d -t -v $(module);)
	go get -d -v ./...

	@echo
	@echo "---> go vet"
	-go vet ./...

	@echo
	@echo "---> Running tests"
	-env NR_LICENSE=123 NQ_URL=$(NQ_URL) routingKey=$(routingKey) exchangeName=$(exchangeName) exchangeType=$(exchangeType) snToken=$(snToken) snHost=$(snHost) subscriptionURL=$(subscriptionURL) go test -v ./... -coverprofile coverage.out -coverpkg ./... > unittest.out
	-cat unittest.out  #for logging to stdout
	-cat unittest.out | go-junit-report > unittest.xml
	-$(GO_ENV) gocov convert coverage.out | gocov-xml > coverage.xml

	@echo

	@echo
	@echo "---> Stopping rabbitmq image"
	@docker stop $(CID2)
	@docker rm $(CID2)

	#check for test failures
	cat unittest.out | grep "^FAIL" && exit 1 || true

# run a code scan
# todo: fix error handling and remove the exclude flag below
.PHONY: scan
scan:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	CGO_ENABLED=0 gosec -exclude=G104 ./...
