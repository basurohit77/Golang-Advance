.PHONY: all
all: static_analysis build test benchmark

.PHONY: build
build:
	@go build

.PHONY: coverage
coverage:
	@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

.PHONY: format
format:
	@find . -type f -name "*.go*" -not -path "./vendor/*" -print0 | xargs -0 gofmt -s -w

# Run all tests
.PHONY: test
test:
	GO111MODULE=on go test -count=1 -v -race -mod=vendor ./... -run ''

# Run only tests that are short
.PHONY:
short_test:
	GO111MODULE=on go test -v -race -mod=vendor ./... -run '' -short

.PHONY:
acceptance:
	GO111MODULE=on ginkgo -v -r

# Run only tests that are Integration test, i.e. with Integration in the name.
.PHONY: integration
integration:
	@go test -v -race ./... -run Integration

.PHONY: static_analysis
static_analysis:
	GO111MODULE=on GOFLAGS='-mod=vendor' golangci-lint run --enable gofmt -E golint -E gosec

# Tools installation
.PHONY: install_tools
install_tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.30.0
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.5.0

.PHONY: benchmark
benchmark:
	@go test -bench=. -benchmem -run=^a

.PHONY: clean
clean:
	@go clean ./...
