.PHONY: all build test lint generate clean install run

# Binary output
BIN := quest-one
CMD := ./cmd/quest-one

# Go flags
CGO_ENABLED := 1
# sqlite_fts5 enables FTS5 full-text search in the mattn/go-sqlite3 driver.
BUILD_TAGS := sqlite_fts5
GOFLAGS := CGO_ENABLED=$(CGO_ENABLED)

all: generate build

## generate: Run code generation (templ → Go)
generate:
	$(GOFLAGS) go generate ./...

## build: Build the binary
build:
	$(GOFLAGS) go build -tags $(BUILD_TAGS) -o $(BIN) $(CMD)

## install: Install to GOBIN
install:
	$(GOFLAGS) go install -tags $(BUILD_TAGS) $(CMD)

## test: Run all tests
test:
	$(GOFLAGS) go test -tags $(BUILD_TAGS) -race -count=1 ./...

## test-domain: Run only domain unit tests (no CGO required)
test-domain:
	go test -race -count=1 ./internal/domain/...

## test-short: Run tests with -short flag (skip integration tests)
test-short:
	$(GOFLAGS) go test -tags $(BUILD_TAGS) -short -race -count=1 ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## run: Start the server
run: build
	QUEST_ONE_LOG=debug ./$(BIN) serve

## clean: Remove built artifacts
clean:
	rm -f $(BIN)
	find . -name '*.templ_gen.go' -not -name '*.go' -delete

## templ: Install the templ code generator
templ:
	go install github.com/a-h/templ/cmd/templ@latest

## deps: Download and tidy modules
deps:
	go mod download
	go mod tidy

## vuln: Check for known vulnerabilities
vuln:
	govulncheck ./...

## build-darwin-universal: Cross-compile Universal Binary for macOS
build-darwin-universal:
	GOOS=darwin GOARCH=amd64 $(GOFLAGS) go build -tags $(BUILD_TAGS) -o $(BIN)-darwin-amd64 $(CMD)
	GOOS=darwin GOARCH=arm64 $(GOFLAGS) go build -tags $(BUILD_TAGS) -o $(BIN)-darwin-arm64 $(CMD)
	lipo -create -output $(BIN)-darwin-universal $(BIN)-darwin-amd64 $(BIN)-darwin-arm64
	rm $(BIN)-darwin-amd64 $(BIN)-darwin-arm64
