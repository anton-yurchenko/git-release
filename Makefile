I := "⚪"
E := "🔴"
BINARY := $(notdir $(CURDIR))
GO_BIN_DIR := $(GOPATH)/bin
OSES := windows linux
ARCHS := amd64

test: lint
	@echo "$(I) unit testing... [this may take a couple of minutes]"
	@go test -v $$(go list ./... | grep -v vendor | grep -v mocks) -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: lint
lint: $(GO_LINTER)
	@echo "$(I) vendoring..."
	@go get ./... || (echo "$(E) 'go get' error"; exit 1)
	@go mod vendor || (echo "$(E) 'go mod vendor' error"; exit 1)
	@go mod tidy || (echo "$(E) 'go mod tidy' error"; exit 1)
	@echo "$(I) linting..."
	@golangci-lint run ./... || (echo "$(E) linter error"; exit 1)

.PHONY: init
init:
	@rm -rf go.mod go.sum ./vendor
	@go mod init $$(pwd | awk -F'/' '{print $$NF}')

.PHONY: build
build: test
	@echo "$(I) building binaries for javascript executor..."
	@rm -rf ./bin
	@mkdir -p bin
	@for ARCH in $(ARCHS); do \
		for OS in $(OSES); do \
			if test "$$OS" = "windows"; then \
				GOOS=$$OS GOARCH=$$ARCH go build -o bin/$(BINARY)-$$OS-$$ARCH.exe; \
			else \
				GOOS=$$OS GOARCH=$$ARCH go build -o bin/$(BINARY)-$$OS-$$ARCH; \
			fi; \
		done; \
	done

.PHONY: codecov
codecov: test
	@go tool cover -html=coverage.txt || (echo "$(E) 'go tool cover' error"; exit 1)

GO_LINTER := $(GO_BIN_DIR)/golangci-lint
$(GO_LINTER):
	@echo "installing linter..."
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint