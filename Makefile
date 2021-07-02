BINARY := $(notdir $(CURDIR))
GO_BIN_DIR := $(GOPATH)/bin
OSES := windows linux
ARCHS := amd64

test: lint
	@echo "unit testing..."
	@go test -v $$(go list ./... | grep -v vendor | grep -v mocks) -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: lint
lint: $(GO_LINTER)
	@echo "vendoring..."
	@go mod vendor
	@go mod tidy
	@echo "linting..."
	@golangci-lint run ./...

.PHONY: init
init:
	@rm -f go.mod
	@rm -f go.sum
	@rm -rf ./vendor
	@go mod init $$(pwd | awk -F'/' '{print "github.com/"$$(NF-1)"/"$$NF}')

# linter
GO_LINTER := $(GO_BIN_DIR)/golangci-lint
$(GO_LINTER):
	@echo "installing linter..."
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: build
build: test
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
	@go tool cover -html=coverage.txt

GO_VER := $$(grep -oE "const Version string = \"[0-9]+.[0-9]+.[0-9]+\"" main.go | tr -d 'const Version string = "')
DOCKER_VER := $$(grep -oE "LABEL \"version\"=\"[0-9]+.[0-9]+.[0-9]+\"" Dockerfile | tr -d 'LABEL "version"="')
JS_VER := $$(jq -r '.version' package.json)

.PHONY: release
release: build
	@if [ "${tag}" != "v${DOCKER_VER}" ] || [ "${tag}" != "v${DOCKER_VER}" ] || [ "${tag}" != "v${JS_VER}" ]; then\
		echo "---> Inconsistent Versioning!";\
		echo "git tag:		${tag}";\
		echo "main.go version:	${GO_VER}";\
		echo "Dockerfile version:	${DOCKER_VER}";\
		echo "package.json version:	${JS_VER}";\
		exit 1;\
	fi
	@git add -A
	@git commit -m $(tag)
	@git push
	@git tag $(tag)
	@git push origin $(tag)