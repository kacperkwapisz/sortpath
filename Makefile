.PHONY: build test clean install release help

# Get version from VERSION file
VERSION := $(shell cat VERSION)
LDFLAGS := -X main.Version=$(VERSION)

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the binary for current platform"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  install  - Install to /usr/local/bin"
	@echo "  release  - Create a new release (requires VERSION bump)"
	@echo "  version  - Show current version"

build:
	go build -ldflags="$(LDFLAGS)" -o bin/sortpath ./cmd/sortpath.go

test:
	go test ./...

clean:
	rm -rf bin/
	rm -rf dist/

install: build
	sudo cp bin/sortpath /usr/local/bin/sortpath

# Cross-compile for all platforms
build-all:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/sortpath-linux-amd64 ./cmd/sortpath.go
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/sortpath-linux-arm64 ./cmd/sortpath.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/sortpath-darwin-amd64 ./cmd/sortpath.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/sortpath-darwin-arm64 ./cmd/sortpath.go
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/sortpath-windows-amd64.exe ./cmd/sortpath.go

# Release management
version:
	@echo "Current version: $(VERSION)"

bump-patch:
	@perl -pe 's/^(\d+)\.(\d+)\.(\d+)$$/sprintf("%d.%d.%d", $$1, $$2, $$3+1)/e' -i VERSION
	@echo "Version bumped to: $$(cat VERSION)"

bump-minor:
	@perl -pe 's/^(\d+)\.(\d+)\.(\d+)$$/sprintf("%d.%d.0", $$1, $$2+1)/e' -i VERSION
	@echo "Version bumped to: $$(cat VERSION)"

bump-major:
	@perl -pe 's/^(\d+)\.(\d+)\.(\d+)$$/sprintf("%d.0.0", $$1+1)/e' -i VERSION
	@echo "Version bumped to: $$(cat VERSION)"

release: build-all
	git add VERSION
	git commit -m "Release v$(VERSION)"
	git tag "v$(VERSION)"
	git push origin main --tags