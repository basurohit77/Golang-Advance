# Added to properly pick up the dependencies.
PRIVATE_MODULES=$(shell grep "^\s*github.ibm.com" go.mod | awk '{print $$1}')
# run unit tests
.PHONY: test
test:
	@docker ps >> /dev/null -- #error if docker not running/not configured
	$(eval PG_DB = pnptest)
	$(eval PG_DB_USER = pnp)
	$(eval PG_DB_PASS = pnp)
	$(eval DB_SSLMODE = disable)
	$(eval PG_DB_PORT = 4321)
	$(eval PG_DB_IP = $(shell ipport=$${DOCKER_HOST#*//};  dbip=$${ipport%%:*}; if [[ $$dbip == "" ]]; then echo localhost; else echo $$dbip; fi))
	# In case the IP is not being discovered during a local run, use this next line instead of the previous
	$(eval RABBITMQ_PORT = 5672)
	$(eval NQ_URL = amqp://guest:guest@$(PG_DB_IP):5672)
	$(eval SNURL = watsontest)
	$(eval SNToken = abc123)
	$(eval SNOW_BYPASS_FLAG = true)
	$(eval BYPASS_LOCAL_STORAGE = true)


	@echo "---> Building postgres image"
	@docker build -f Dockerfile.local -t pnp-nq2ds-test-pg .
	@echo
	@echo "---> Building rabbitmq image"
	@docker build -f Dockerfile.rabbitmq -t pnp-rabbitmq .
	@echo
	@echo "---> Starting postgres image"
	$(eval CID1 = $(shell docker run -d -p $(PG_DB_PORT):5432 pnp-nq2ds-test-pg))
	@sleep 5
	@echo
	@echo "---> Starting rabbitmq image"
	$(eval CID2 = $(shell docker run -d -p $(RABBITMQ_PORT):5672 --hostname localhost pnp-rabbitmq))

	@echo
	@echo "---> Installing dependencies"
    # Added to properly pick up the dependencies.
	$(shell go env -w GOPRIVATE="*.ibm.com")
	$(foreach module,$(PRIVATE_MODULES), go get -d -t -v $(module);)
	go get -d -v ./...

	@echo
	@echo "---> Running tests"
	-env PG_DB_USER=$(PG_DB_USER) PG_DB=$(PG_DB) PG_DB_IP=$(PG_DB_IP) PG_DB_PASS=$(PG_DB_PASS) PG_DB_PORT=$(PG_DB_PORT) DB_SSLMODE=$(DB_SSLMODE) NQ_URL=$(NQ_URL) SNURL=$(SNURL) SNToken=$(SNToken) SNOW_BYPASS_FLAG=$(SNOW_BYPASS_FLAG) BYPASS_LOCAL_STORAGE=$(BYPASS_LOCAL_STORAGE) go test -v `go list ./... | grep -v testutils` -coverprofile coverage.out -coverpkg ./... > unittest.out
	-cat unittest.out  #for logging to stdout

	@echo
	@echo "---> Stopping postgres image"
	@docker stop $(CID1)
	@echo "---> Removing postgres image"
	@docker rm $(CID1)
	@echo
	@echo "---> Stopping rabbitmq image"
	@docker stop $(CID2)
	@echo "---> Removing rabbitmq image"
	@docker rm $(CID2)

	#check for test failures
	cat unittest.out | grep "^FAIL" && exit 1 || true

# run a code scan
# todo: fix error handling and remove the exclude flag below
.PHONY: scan
scan:
	go get -u github.com/securego/gosec/v2/cmd/gosec
	CGO_ENABLED=0 gosec -exclude=G104 ./...
