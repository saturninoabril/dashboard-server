# Docker Build Versions
DOCKER_BUILDER_SERVER_IMAGE = golang:1.16
DOCKER_BUILDER_WEBAPP_IMAGE = node:16-alpine
DOCKER_BASE_IMAGE = alpine:3.12

export GOBIN ?= $(PWD)/bin
GO ?= $(shell command -v go 2> /dev/null)
GOFLAGS ?= $(GOFLAGS:)
GO_CONTROLLER_IMAGE ?= saturninoabril/dashboard-server:test
BUILDER_GOOS_GOARCH="$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)"

# Build Flags
BUILD_NUMBER ?= $(BUILD_NUMBER:)
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH = $(shell git rev-parse HEAD)
# If we don't set the build number it defaults to dev
ifeq ($(BUILD_NUMBER),)
	BUILD_NUMBER := dev
endif

LDFLAGS += -X "github.com/saturninoabril/dashboard-server/model.BuildNumber=$(BUILD_NUMBER)"
LDFLAGS += -X "github.com/saturninoabril/dashboard-server/model.BuildTime=$(BUILD_TIME)"
LDFLAGS += -X "github.com/saturninoabril/dashboard-server/model.BuildHash=$(BUILD_HASH)"

## Checks the code style, tests, builds and bundles.
all: check-style

## Build the docker image for server
.PHONY: build-image
build-image:
	@echo Building Go Controller Container Image
	docker build \
	--build-arg DOCKER_BUILDER_SERVER_IMAGE=$(DOCKER_BUILDER_SERVER_IMAGE) \
	--build-arg DOCKER_BUILDER_WEBAPP_IMAGE=$(DOCKER_BUILDER_WEBAPP_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	--build-arg ENV=prod \
	--build-arg GITHUB_USERNAME=$(GITHUB_USERNAME) \
	--build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) \
	. -t $(GO_CONTROLLER_IMAGE) \
	--no-cache

DASHBOARD_DATABASE_DEV ?= postgres://dashboarduser:dashboardpwd@localhost:5433/dashboard_dev?sslmode=disable
export DASHBOARD_DATABASE=${DASHBOARD_DATABASE_DEV}

DASHBOARD_DATABASE_TEST ?= postgres://dashboarduser:dashboardpwd@localhost:5433/dashboard_test?sslmode=disable

## Build the Container image for local development.
.PHONY: start
start:
	@echo Starting local development
	docker compose up -d
	sleep 10
	dashboard server --debug --dev

## Shutdown the development environment.
.PHONY: stop
stop:
	@echo Shutting down the local environment
	docker compose down

## Clean the development environment.
.PHONY: clean
clean:
	@echo Cleaning the local environment
	docker compose rm -s

## Runs govet and gofmt against all packages.
.PHONY: check-style
check-style: govet lint
	@echo Checking for style guide compliance

## Runs lint against all packages.
.PHONY: lint
lint:
	@echo Running lint
	$(GO) get -u golang.org/x/lint/golint
	$(GOBIN)/golint -set_exit_status
	@echo lint success

## Runs govet against all packages.
.PHONY: vet
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

build-linux:
	@echo Build Linux amd64
ifeq ($(BUILDER_GOOS_GOARCH),"linux_amd64")
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o $(GOBIN) ./...
else
	mkdir -p $(GOBIN)/linux_amd64
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o $(GOBIN)/linux_amd64 ./...
endif

build-osx:
	@echo Build OSX amd64
ifeq ($(BUILDER_GOOS_GOARCH),"darwin_amd64")
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o $(GOBIN) ./...
else
	mkdir -p $(GOBIN)/darwin_amd64
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o $(GOBIN)/darwin_amd64 ./...
endif

.PHONY: build
build: build-linux build-osx

### Tests

TESTS=.

.PHONY: go-junit-report
go-junit-report:
	$(GO) get github.com/jstemmer/go-junit-report

.PHONY: test-server
test-server: go-junit-report
	@echo Running server tests
	DASHBOARD_DATABASE_TEST=${DASHBOARD_DATABASE_TEST} DASHBOARD_TABLE_PREFIX=dashboard_ $(GO) test $(TESTFLAGS) -v -tags 'e2e'  -covermode=count -coverprofile=coverage.out ./...
	EXIT_STATUS=$?
	cat output | $(GOBIN)/go-junit-report > report.xml
	rm output
	exit $EXIT_STATUS

generate:
	go get -modfile=go.tools.mod github.com/jteeuwen/go-bindata
	# go install -v github.com/jteeuwen/go-bindata/...
	go generate ./...
