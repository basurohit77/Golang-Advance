.PHONY: all
all: lint build test benchmark

.PHONY: build
build:
	@go build

.PHONY: coverage
coverage:
	@go test -coverprofile=coverage.out && go tool cover -html=coverage.out

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
	GO111MODULE=on -v -r

# Run only tests that are Integration test, i.e. with Integration in the name.
.PHONY: integration
integration:
	@go test -v -race ./... -run Integration

# Linting the code using golangci-lint, which must be installed firt.
.PHONY: lint
lint:
	GO111MODULE=on golangci-lint run --enable gofmt

.PHONY: benchmark
benchmark:
	@go test -bench=. -benchmem -run=^a

.PHONY: clean
clean:
	@go clean ./...
