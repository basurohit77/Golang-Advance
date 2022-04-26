.PHONY: all
all: lint build test

.PHONY: build
build:
	GO111MODULE=on go build .

.PHONY: coverage
coverage:
	GO111MODULE=on go test -coverprofile=coverage.out ./... -run '' && go tool cover -html=coverage.out

.PHONY: format
format:
	@find . -type f -name "*.go*" -not -path "./vendor/*" -print0 | xargs -0 gofmt -s -w

.PHONY: test
test:
	GO111MODULE=on go test -count=1 -v -race ./... -run ''

# In case you have not installed golangci-lint, this is a make target to install it.
.PHONY: lint-deps
lint-deps:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.30.0

# Linting the code using golangci-lint, which must be installed firt.
.PHONY: lint
lint:
	GO111MODULE=on golangci-lint run --enable gofmt -E golint --exclude-use-default=false

.PHONY: benchmark
benchmark:
	GO111MODULE=on go test -bench=. -benchmem ./... -run=^a -benchtime=10s
