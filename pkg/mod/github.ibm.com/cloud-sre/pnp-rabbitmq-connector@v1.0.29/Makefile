# some useful variables
VERBOSE =

#for local development to test
.PHONY: test
test: cicd-setup deps scan unit-test clean

# download dependencies for binary
.PHONY: deps
deps:
	go get -d $(VERBOSE)

# run unit tests
.PHONY: unit-test
unit-test:
	@docker ps >> /dev/null -- #error if docker not running/not configured
	$(eval DOCKER_IP = $(shell ipport=$${DOCKER_HOST#*//};  dbip=$${ipport%%:*}; if [[ $$dbip == "" ]]; then echo localhost; else echo $$dbip; fi))
	@echo
	@echo "---> Running tests"
	go vet -v ./...
	env NQ_URL=amqp://guest:guest@$(DOCKER_IP):5672 NR_LICENSE=123 go test -v ./... -coverpkg ./... -coverprofile coverage.out ./...

# run a code scan
.PHONY: scan
scan:
	go get -u github.com/securego/gosec/v2/cmd/gosec
	CGO_ENABLED=0 gosec ./...

.PHONY: cicd-setup
cicd-setup:
	$(eval VERBOSE = -v)
	./cicd-setup.sh || true

.PHONY: clean
clean:
	./cicd-cleanup.sh || true

