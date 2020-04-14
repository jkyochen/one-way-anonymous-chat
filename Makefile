SERVICE ?=one-way-anonymous-chat

GO ?= go
GOFMT ?= gofmt "-s"
PACKAGES ?= $(shell $(GO) list ./... | grep -v integrations)
GOFILES := $(shell find . -name "*.go" -type f)
TAGS ?= sqlite sqlite_unlock_notify json1

ifneq ($(shell uname), Darwin)
	EXTLDFLAGS = -extldflags "-static" $(null)
else
	EXTLDFLAGS =
endif

all: build

.PHONY: run
run:
	$(GO) run cmd/*

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: test
test: fmt-check
	@$(GO) test -v -cover -tags '$(TAGS)' -coverprofile coverage.txt $(PACKAGES) && echo "\n==>\033[32m Ok\033[m\n" || exit 1

.PHONY: lint
lint:
	@hash revive > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mgechev/revive; \
	fi
	revive -config .revive.toml ./... || exit 1

.PHONY: golangci-lint
golangci-lint:
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		export BINARY="golangci-lint"; \
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin v1.23.6; \
	fi
	golangci-lint run --deadline=3m

build: $(SERVICE)

$(SERVICE): $(GOFILES)
	$(GO) build -v -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s' -o bin/$@ ./cmd/*

.PHONY: check
check: test lint golangci-lint
