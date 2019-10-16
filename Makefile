SHELL = /bin/sh

ifneq ("$(wildcard .env)","")
	include .env
	export
endif

APP_NAME = iot
PACKAGES ?= ./...
GO_FILES ?= */*/*.go

OUTPUR_DIR=bin
BINARY_PATH=$(OUTPUR_DIR)/$(APP_NAME)

SERVER_SOURCE = cmd/iot/iot.go
SERVER_RUNNER = go run $(SERVER_SOURCE)
ifeq ($(DEBUG), true)
	SERVER_RUNNER = dlv debug $(SERVER_SOURCE) --
endif

WORKER_SOURCE = cmd/worker/worker.go
WORKER_RUNNER = go run $(WORKER_SOURCE)
ifeq ($(DEBUG), true)
	WORKER_RUNNER = dlv debug $(WORKER_SOURCE) --
endif

.DEFAULT_GOAL := app

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## name: Output app name
.PHONY: name
name:
	@echo -n $(APP_NAME)

## version: Output last commit sha1
.PHONY: version
version:
	@echo -n $(shell git rev-parse --short HEAD)

## app: Build app with dependencies download
.PHONY: app
app: deps go

## go: Build app
.PHONY: go
go: format lint test bench build

## deps: Download dependencies
.PHONY: deps
deps:
	go get github.com/kisielk/errcheck
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports

## format: Format code
.PHONY: format
format:
	goimports -w $(GO_FILES)
	gofmt -s -w $(GO_FILES)

## lint: Lint code
.PHONY: lint
lint:
	golint $(PACKAGES)
	errcheck -ignoretests $(PACKAGES)
	go vet $(PACKAGES)

## test: Test with coverage
.PHONY: test
test:
	script/coverage

## bench: Benchmark code
.PHONY: bench
bench:
	go test $(PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH) $(SERVER_SOURCE)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH)-worker $(WORKER_SOURCE)

## build: Build binary for ARM
.PHONY: build-arm
build-arm:
	CGO_ENABLED=0 GOARCH="arm" GOOS="linux" GOARM="6" go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH)-arm $(SERVER_SOURCE)
	CGO_ENABLED=0 GOARCH="arm" GOOS="linux" GOARM="6" go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH)-arm-worker $(WORKER_SOURCE)

## start-worker: Start worker
.PHONY: start-worker
start-worker:
	$(WORKER_RUNNER) \
		-mqttClientID "iot-worker-dev" \
		-mqttServer $(IOT_MQTT_SERVER) \
		-mqttPort $(IOT_MQTT_PORT) \
		-mqttUser $(IOT_MQTT_USER) \
		-mqttPass $(IOT_MQTT_PASS) \
		-subscribe "dev-worker" \
		-publish "dev" \
		-hueUsername $(BRIDGE_USERNAME) \
		-hueBridgeIP $(BRIDGE_IP) \
		-netatmoAccessToken "$(NETATMO_ACCESS_TOKEN)" \
		-netatmoClientID "$(NETATMO_CLIENT_ID)" \
		-netatmoClientSecret "$(NETATMO_CLIENT_SECRET)" \
		-netatmoRefreshToken "$(NETATMO_REFRESH_TOKEN)" \
		-sonosAccessToken "$(SONOS_ACCESS_TOKEN)" \
		-sonosClientID "$(SONOS_CLIENT_ID)" \
		-sonosClientSecret "$(SONOS_CLIENT_SECRET)" \
		-sonosRefreshToken "$(SONOS_REFRESH_TOKEN)"

## start: Start app
.PHONY: start
start:
	$(SERVER_RUNNER) \
		-mqttClientID "iot-dev" \
		-mqttServer $(IOT_MQTT_SERVER) \
		-mqttPort $(IOT_MQTT_PORT) \
		-mqttUser $(IOT_MQTT_USER) \
		-mqttPass $(IOT_MQTT_PASS) \
		-subscribe "dev" \
		-publish "dev-worker" \
		-prometheus \
		-csp "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'"
