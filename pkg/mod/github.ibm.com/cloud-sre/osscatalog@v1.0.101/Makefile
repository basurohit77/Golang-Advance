VERBOSE =
GO_ENV =
GO_DOCKER_ENV = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
IMAGE_NAME = $(shell jq -r .component_name metadata.json || sed -n -e 's/^.*appName://p' Jenkinsfile | awk -F\' '{print $$2}')
# eventually we should switch to separate registries for dev and stage/prod and limit access to the latter
IMAGE_REGISTRY = wcp-cto-oss-docker-local.artifactory.swg-devops.com

MAJOR_GOVERSION=$(shell go version | grep -o '1\.[0-99]*')
PACKAGES=$(shell find ./ -type d -not -path '*/\.*')
PRIVATE_MODULES=$(shell grep "^\s*github.ibm.com\/\(cloud-sre\|sosat\)" go.mod | awk '{print $$1}')
SINGLE_MODULE=$(shell grep "^require\s*github.ibm.com\/\(cloud-sre\|sosat\)" go.mod | awk '{print $$2}') # for modules that require one package
PRIVATE_MODULES += $(SINGLE_MODULE)
CURRENT_MAKEFILE:=$(lastword $(MAKEFILE_LIST))
SELF_DIR:=$(shell pwd)

# create binary
.PHONY: quick-local-binary
quick-local-binary:
	go build -o osscatimporter ./cmd/osscatimporter
	go build -o osscatpublisher ./cmd/osscatpublisher

.PHONY: binary
binary:
	$(GO_DOCKER_ENV) go build $(VERBOSE) -o osscatimporter ./cmd/osscatimporter
	$(GO_DOCKER_ENV) go build $(VERBOSE) -o osscatpublisher ./cmd/osscatpublisher


# download dependencies for binary
.PHONY: deps
deps:
	echo "GO version: $(MAJOR_GOVERSION)"
	$(shell go env -w GO111MODULE=on)
	$(shell go env -w GOPRIVATE="*.ibm.com")
	$(shell go env -w GONOPROXY="none")
	$(shell go env -w GONOSUMDB="*.ibm.com")

	$(foreach module,$(PRIVATE_MODULES),$(GO_ENV) go get -d -t $(VERBOSE) $(module);)
	$(GO_ENV) go get -d $(VERBOSE) ./...


# run unit tests
# get dependencies for gocov-xml and go-junit-report so we can send test results to devops-insights
.PHONY: test
test:
	$(GO_ENV) go get github.com/axw/gocov/gocov
	$(GO_ENV) go get github.com/AlekSi/gocov-xml
	$(GO_ENV) go get -u github.com/jstemmer/go-junit-report

	$(foreach module,$(PRIVATE_MODULES),$(GO_ENV) go get -d -t $(VERBOSE) $(module);)
	$(GO_ENV) go get -d -t $(VERBOSE) ./...
	$(GO_ENV) go vet ./...

	echo "GO version: $(MAJOR_GOVERSION)"

	-$(GO_ENV) go test `go list ./... | grep -v testutils` $(VERBOSE) -short -coverprofile coverage-all.out  > unittest.out
	-cat unittest.out  #for logging to stdout
	-cat unittest.out | go-junit-report > unittest.xml
	-$(GO_ENV) gocov convert coverage-all.out | gocov-xml > coverage.xml

	-cp unittest.xml build/reports/test-results.xml
	-cp coverage.xml build/reports/coverage.xml

	cat unittest.out | grep "^FAIL" && exit 1 || true

# run a code scan
# todo: fix error handling and remove the exclude flag below
.PHONY: scan
scan:
	$(GO_ENV) go install github.com/securego/gosec/v2/cmd/gosec@v2.9.6
	CGO_ENABLED=0 gosec -exclude=G104 ./...


# build the docker image
.PHONY: image
image: binary
	docker build -t $(IMAGE_NAME):latest .

# tag and push official image
.PHONY: image-push
image-push:
	git diff-index --quiet HEAD -- #error if uncommitted changes
	$(eval IMAGE_TAG = $(shell git rev-parse HEAD))
	docker tag $(IMAGE_NAME):latest $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)


# push image and deploy to OSSDev personal namespace
# todo: run kdep with custom tag
.PHONY: deploy
deploy:
	$(eval IMAGE_TAG = $(shell whoami|cut -d@ -f1)-$(shell date +%s))
	docker tag $(IMAGE_NAME):latest $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)


# target ran by Jenkins pipeline for buildPublishComponent
.PHONY: cicd-full
cicd-full: cicd-full-setup deps test scan image clean

# target ran by Jenkins pipeline for buildDependentComponents
.PHONY: cicd-test
cicd-test: cicd-full-setup deps test scan clean

.PHONY: cicd-full-setup
cicd-full-setup:
	$(eval VERBOSE = -v)
	$(eval GO_ENV = $(GO_DOCKER_ENV))
	$(eval GO111MODULE=on)
	./cicd-setup.sh || true

# build the binary and publishes it to the developer's namespace in OSSDev, in the future also calls kdep to deploy
.PHONY: cicd
cicd: image deploy


# remove artifacts and cleanup
.PHONY: clean
clean:
	rm main || true
	./cicd-cleanup.sh || true


#
# Initialize
#

init:
	@echo 'Getting build scripts....'
	git clone git@github.ibm.com:cloud-platform-sre/sre-common-build-scripts.git

ci-setup:
	$(eval VERBOSE = -v)
	$(eval GO_ENV = $(GO_DOCKER_ENV))
	$(eval GO111MODULE=on)
	${eval GOCACHE=/tmp/}

#
# Clean up
#

clean-all: clean-reports

clean-reports: check-common-build-scripts
	$(RM) -r .coverage
	${SELF_DIR}/sre-common-build-scripts/scripts/clean-reports.sh

clean-hadolint-report:
	@rm -rf build/reports/dockerlint.xml

clean-ci:
	-rm -rf ${SELF_DIR}/sre-common-build-scripts
	-rm -rf cocoa
	-rm -rf build
	-rm -rf code-deployable-artifacts
	-rm checkImageUbiOrScratch.sh signing_pfx_file.pfx
	-rm unittest.out unittest.xml coverage.xml coverage-all.out

#
# Check code
#

ensure-reports:
	@mkdir -p build/reports

ensure-build:
	@mkdir -p build

lint: lint-dockerfile

lint-report: lint-dockerfile-report

lint-dockerfile: check-common-build-scripts
	@./sre-common-build-scripts/scripts/docker-lint.sh -d ./Dockerfile

lint-dockerfile-report: check-common-build-scripts ensure-reports
	@./sre-common-build-scripts/scripts/docker-lint.sh -d ./Dockerfile -f build/reports/dockerlint.xml || (echo 'Error: Dockerfile verification failed, refer to: ${BUILD_URL}dockerlint'; exit 1)

go-scan: ensure-reports
	go install github.com/securego/gosec/v2/cmd/gosec@v2.9.6
	CGO_ENABLED=0 gosec -exclude=G104 -fmt=html -out=build/reports/sourcelint.html -stdout -verbose=text ./...

#
# Pipeline
#

pipeline-init: init clean-reports

pipeline-static-code-analysis: deps go-scan

pipeline-static-code-lint-analysis: lint-report deps go-scan

pipeline-tests: ensure-reports ci-setup deps test

pipeline-build: ci-setup deps binary

pipeline-publish-pr:
	@echo "Not publishing any artifacts in PR pipeline"

pipeline-publish-merge:
	@echo "Not publishing any artifacts in Merge pipeline"

pipeline-cleanup: clean-ci

#
# Utilities
#

check-common-build-scripts:
	@if ! test -d ${SELF_DIR}/sre-common-build-scripts; then \
		echo "${SELF_DIR}/sre-common-build-scripts does not exist. Run the 'make -f ${CURRENT_MAKEFILE} init' command to create it."; \
		false; \
	fi
