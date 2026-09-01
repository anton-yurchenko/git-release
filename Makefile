I := "⚪"
E := "🔴"
BINARY := $(notdir $(CURDIR))
OSES := windows linux
ARCHS := amd64

GO_BIN_DIR := $(shell go env GOPATH)/bin
# NOTE: tools must be built with the toolchain this module requires, otherwise
# they can not type-check it ("package requires newer Go version")
GOTOOLCHAIN_PIN := $(shell awk '/^go /{print "go"$$2; exit}' go.mod)
# NOTE: keep in sync with the golangci-lint-action version in .github/workflows/test.yml
GOLANGCI_LINT_VERSION := v2.13.2
MOCKERY_VERSION := v3.7.4
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || echo $(GO_BIN_DIR)/golangci-lint)
MOCKERY := $(shell command -v mockery 2>/dev/null || echo $(GO_BIN_DIR)/mockery)

.PHONY: all
all: test

.PHONY: tools
tools:
	@echo "$(I) installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@GOTOOLCHAIN=$(GOTOOLCHAIN_PIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) || (echo "$(E) 'go install golangci-lint' error"; exit 1)
	@echo "$(I) installing mockery $(MOCKERY_VERSION)..."
	@GOTOOLCHAIN=$(GOTOOLCHAIN_PIN) go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION) || (echo "$(E) 'go install mockery' error"; exit 1)

.PHONY: tidy
tidy:
	@echo "$(I) tidying go modules..."
	@go mod tidy || (echo "$(E) 'go mod tidy' error"; exit 1)

.PHONY: mocks
mocks:
	@test -x "$(MOCKERY)" || (echo "$(E) mockery not found, run 'make tools'"; exit 1)
	@echo "$(I) generating mocks..."
	@$(MOCKERY) || (echo "$(E) 'mockery' error"; exit 1)

.PHONY: lint
lint:
	@test -x "$(GOLANGCI_LINT)" || (echo "$(E) golangci-lint not found, run 'make tools'"; exit 1)
	@echo "$(I) linting..."
	@$(GOLANGCI_LINT) run ./... || (echo "$(E) linter error"; exit 1)

.PHONY: test
test: lint
	@echo "$(I) unit testing... [this may take a couple of minutes]"
	@go test -v $$(go list ./... | grep -v '/mocks$$') -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: build
build: test
	@echo "$(I) building binaries for javascript executor..."
	@rm -rf ./bin
	@mkdir -p bin
	@for ARCH in $(ARCHS); do \
		for OS in $(OSES); do \
			if test "$$OS" = "windows"; then \
				CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build -trimpath -buildvcs=false -ldflags="-w -s" -o bin/$(BINARY)-$$OS-$$ARCH.exe; \
			else \
				CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build -trimpath -buildvcs=false -ldflags="-w -s" -o bin/$(BINARY)-$$OS-$$ARCH; \
			fi; \
		done; \
	done

.PHONY: codecov
codecov: test
	@go tool cover -html=coverage.txt || (echo "$(E) 'go tool cover' error"; exit 1)
