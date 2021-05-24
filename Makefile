# Docker Build Versions
DOCKER_BUILDER_SERVER_IMAGE = golang:1.16
DOCKER_BUILDER_WEBAPP_IMAGE = node:16-alpine
DOCKER_BASE_IMAGE = alpine:3.12

export GOBIN ?= $(PWD)/bin
GO ?= $(shell command -v go 2> /dev/null)
GOFLAGS ?= $(GOFLAGS:)
GO_CONTROLLER_IMAGE ?= saturninoabril/dashboard-server:test
BUILDER_GOOS_GOARCH="$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)"

MACHINE = $(shell uname -m)
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH = $(shell git rev-parse HEAD)
LDFLAGS += -X "github.com/saturninoabril/dashboard-server/model.BuildHash=$(BUILD_HASH)"

## Checks the code style, tests, builds and bundles.
all: check-style

## Build the docker image for dashboard
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

DASHBOARD_DATABASE_TEST ?= postgres://dashboarduser:dashboardpwd@localhost:5433/dashboard_test?sslmode=disable
export DASHBOARD_DATABASE=${DASHBOARD_DATABASE_TEST}

## Build the Container image for local development.
.PHONY: start
start:
	@echo Starting local development
	docker compose up -d
	sleep 10
	dashboard server --debug --dev

.PHONY: migrate
migrate:
	@echo Creating/migrating schema for local development
	dashboard schema migrate --dev

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
check-style: govet lint check-style-webapp
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

## Runs tests against webapp.
.PHONY: node_modules
test-webapp: node_modules ## Checks JS file for ESLint conformity
	@echo Testing Webapp
	export REACT_APP_BUILD_CLOUD=true
	cd webapp; npm run test

## Runs checks against webapp.
.PHONY: check-style-weabpp
check-style-webapp: node_modules ## Checks JS file for ESLint conformity
	@echo Checking for style guide compliance
	cd webapp; npm run check

fix-style-webapp: node_modules ## Fix JS file ESLint issues
	@echo Fixing lint issues to follow style guide

	cd webapp; npm run fix

## Install node modules.
.PHONY: node_modules
node_modules:
	@echo Getting dependencies using npm
	cd webapp; npm install

### Tests

TESTS=.

.PHONE: go-junit-report
go-junit-report:
	$(GO) get github.com/jstemmer/go-junit-report

.PHONY: test-server
test-server: go-junit-report
	@echo Running all tests
	DASHBOARD_DATABASE=$(DASHBOARD_DATABASE) ./scripts/test-server.sh "$(GO)" "$(GOBIN)" "$(TESTS)" "$(TESTFLAGS)"

