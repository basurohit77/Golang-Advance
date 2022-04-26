# some useful variables
VERBOSE =

#for local development to test
.PHONY: test
test: cicd-setup deps scan unit-test

# download dependencies for binary
.PHONY: deps
deps:
	go get -d $(VERBOSE) ./...

# run unit tests
.PHONY: unit-test
unit-test:
	@echo
	@echo "---> Running tests"
	go vet ./...

	go test `go list ./... | grep -v testutils |sort` -v -coverprofile coverage.out -coverpkg ./... > unittest.out


# run a code scan
.PHONY: scan
scan:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	CGO_ENABLED=0 gosec  ./...

.PHONY: cicd-setup
cicd-setup:
	$(eval VERBOSE = -v)
	./cicd-setup.sh || true
