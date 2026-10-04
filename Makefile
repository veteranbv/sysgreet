.PHONY: fmt lint test test-verbose test-race test-coverage bench vuln build clean help

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

lint:
	go vet ./...
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. See https://golangci-lint.run/welcome/install/" && exit 1)
	golangci-lint run

test:
	CGO_ENABLED=0 go test ./...

test-verbose:
	CGO_ENABLED=0 go test -v ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

test-coverage:
	CGO_ENABLED=0 go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

bench:
	CGO_ENABLED=0 go test -run '^$$' -bench . -benchmem ./test/benchmarks

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

build:
	CGO_ENABLED=0 go build -o sysgreet ./cmd/sysgreet

clean:
	rm -f sysgreet coverage.out

help:
	@echo "Available targets:"
	@echo "  fmt            - Format all Go files"
	@echo "  lint           - Run go vet and golangci-lint"
	@echo "  test           - Run all tests"
	@echo "  test-verbose   - Run all tests with verbose output"
	@echo "  test-race      - Run tests with the race detector (needs cgo)"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  bench          - Run performance benchmarks"
	@echo "  vuln           - Check dependencies and the Go release for known vulnerabilities"
	@echo "  build          - Build the sysgreet binary"
	@echo "  clean          - Remove build artifacts"
	@echo "  help           - Show this help message"
