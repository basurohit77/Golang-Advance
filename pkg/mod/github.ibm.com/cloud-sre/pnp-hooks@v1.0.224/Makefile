.PHONY: cicd
cicd:
	git diff-index --quiet HEAD -- #error if uncommitted changes
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main
	$(eval imageRegistry = wcp-cto-oss-docker-local.artifactory.swg-devops.com/api-pnp-hooks)
	$(eval imageTag = $(shell git rev-parse HEAD)-$(shell date +%s)))
	$(eval imageURI = "$(imageRegistry):$(imageTag)")
	@echo $(imageURI)
	docker build -t $(imageURI) .
	rm main
	docker push $(imageURI)

.PHONY: test
test:
	@docker ps >> /dev/null -- #error if docker not running/not configured
	$(eval RABBITMQ_EXCHANGE_NAME = "pnp.direct")
	$(eval RABBITMQ_EXCHANGE_TYPE = direct)
	$(eval RABBITMQ_TEST_QKEY = "health.hooks:health")
	$(eval API_CATALOG_HEALTHZ_URL = "https://pnp-api-oss.cloud.ibm.com/catalog/api/catalog/healthz")
	$(eval SNOW_TOKEN = apikey)
	$(eval MASTER_KEY = 000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000)
	$(eval DOCKER_HOST_IP = $(shell ipport=$${DOCKER_HOST#*//};  dbip=$${ipport%%:*}; if [[ $$dbip == "" ]]; then echo localhost; else echo $$dbip; fi))
	$(eval RABBITMQ_URL = $(shell echo "amqp://guest:guest@$(DOCKER_HOST_IP):5672"))

	@echo
	./cicd-setup.sh

	@echo
	@echo "---> Installing dependencies"
	go get -d -v ./...

	@echo
	@echo "---> go vet"
	-go vet ./...

	@echo
	@echo "---> Running tests"
	-env RABBITMQ_EXCHANGE_NAME=$(RABBITMQ_EXCHANGE_NAME) RABBITMQ_EXCHANGE_TYPE=$(RABBITMQ_EXCHANGE_TYPE) RABBITMQ_TEST_QKEY=$(RABBITMQ_TEST_QKEY) API_CATALOG_HEALTHZ_URL=$(API_CATALOG_HEALTHZ_URL) SNOW_TOKEN=$(SNOW_TOKEN) MASTER_KEY=$(MASTER_KEY) RABBITMQ_URL=$(RABBITMQ_URL) NR_LICENSE=123 go test -v ./... -coverpkg ./... -coverprofile coverage.out ./...  > unittest.out
	-cat unittest.out  #for logging to stdout

	@echo
	./cicd-cleanup.sh

	#check for test failures
	cat unittest.out | grep "^FAIL" && exit 1 || true

# run a code scan
# todo: fix error handling and remove the exclude flag below
.PHONY: scan
scan:
	go get -u github.com/securego/gosec/v2/cmd/gosec
	CGO_ENABLED=0 gosec -exclude=G104 ./...

