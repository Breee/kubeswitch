BINARY   := kubeswitch
GO       := go
GOFLAGS  := -ldflags="-s -w"
PLATFORMS := linux/amd64 darwin/amd64 darwin/arm64

.PHONY: all build test vet e2e clean install-completions dist

## all: vet, test, build (default target)
all: vet test build

## build: compile the binary for the current platform
build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o $(BINARY)

## test: run unit tests
test:
	$(GO) test -v -count=1 ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## e2e: run end-to-end tests (requires kind + kubectl)
e2e: build
	./e2e_test.sh

## dist: cross-compile and package release tarballs
dist: clean
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		echo "Building $$os/$$arch..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) -o $(BINARY); \
		tar czf $(BINARY)_$${os}_$${arch}.tar.gz $(BINARY) LICENSE README.md; \
	done

## install-completions: install bash completions to user profile
install-completions:
	@echo 'Add the following line to your ~/.bashrc or ~/.bash_profile:'
	@echo '  source $(CURDIR)/completions/kubeswitch.bash'

## clean: remove build artifacts
clean:
	rm -f $(BINARY) $(BINARY)_*.tar.gz

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/^## //' | column -t -s ':'
