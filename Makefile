# Makefile for common development tasks

.PHONY: fmt lint test ci

# Format imports and code. Requires goimports to be installed.
fmt:
	gofmt -w .
	goimports -w -local github.com/lugatuic/goberus .

# Run linters using golangci-lint (expects .golangci.yml/.golangci.yaml configured).
lint:
	golangci-lint run

# Run unit tests
test:
	go test ./...

# CI convenience target: format, lint, test
ci: fmt lint test
