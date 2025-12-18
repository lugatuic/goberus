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

# Print current version
version:
	@cat VERSION

# Bump version (usage: make bump-version VERSION=0.0.2)
bump-version:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make bump-version VERSION=0.0.X"; exit 1; fi
	@echo "$(VERSION)" > VERSION
	@echo "Version bumped to $(VERSION)"

# Release target: creates a tag and pushes to remote (requires VERSION to be set)
release: bump-version
	@VERSION=$$(cat VERSION) && \
		echo "Releasing v$$VERSION..." && \
		git add VERSION CHANGELOG.md && \
		git commit -m "release: v$$VERSION" && \
		git tag -a "v$$VERSION" -m "Release v$$VERSION" && \
		echo "Tag v$$VERSION created. Run 'git push origin v$$VERSION' to publish."
