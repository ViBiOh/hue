SHELL = /bin/sh

APP_NAME ?= iot
VERSION ?= $(shell git rev-parse --short HEAD)
AUTHOR ?= $(shell git log --pretty=format:'%an' -n 1)

PACKAGES ?= ./...
APP_PACKAGES = $(shell go list -e $(PACKAGES) | grep -v vendor | grep -v node_modules)

GOBIN=bin
BINARY_PATH=$(GOBIN)/$(APP_NAME)

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

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## $(APP_NAME): Build app with dependencies download
$(APP_NAME): deps go

.PHONY: go
go: format lint test bench build

## name: Output name
.PHONY: name
name:
	@echo -n $(APP_NAME)

## dist: Output build output path
.PHONY: dist
dist:
	@echo -n $(BINARY_PATH)

## version: Output sha1 of last commit
.PHONY: version
version:
	@echo -n $(VERSION)

## author: Output author's name of last commit
.PHONY: author
author:
	@python -c 'import sys; import urllib; sys.stdout.write(urllib.quote_plus(sys.argv[1]))' "$(AUTHOR)"

## deps: Download dependencies
.PHONY: deps
deps:
	go get github.com/golang/dep/cmd/dep
	go get github.com/kisielk/errcheck
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports
	dep ensure

## format: Format code
.PHONY: format
format:
	goimports -w */*/*.go
	gofmt -s -w */*/*.go

## lint: Lint code
.PHONY: lint
lint:
	golint $(APP_PACKAGES)
	errcheck -ignoretests $(APP_PACKAGES)
	go vet $(APP_PACKAGES)

## test: Test code with coverage
.PHONY: test
test:
	script/coverage

## bench: Benchmark code
.PHONY: bench
bench:
	go test $(APP_PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH) $(SERVER_SOURCE)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH)-worker $(WORKER_SOURCE)

## install: Install binary in GOPATH
.PHONY: install
install:
	go install github.com/ViBiOh/iot/cmd/iot
	go install github.com/ViBiOh/iot/cmd/worker

## systemd: Configure systemd for launching local and remote worker
.PHONY: systemd
systemd:
	sudo cp systemd/* /lib/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable iot-local.service iot-local-worker.service
	sudo systemctl restart iot-local.service iot-local-worker.service

## update-worker: Update worker by fetching new version and restarting services
.PHONY: update-worker
update-worker: deps install systemd

## start-worker: Start worker
.PHONY: start-worker
start-worker:
	$(WORKER_RUNNER) \
		-dbHost $(IOT_DATABASE_HOST) \
		-dbName $(IOT_DATABASE_NAME) \
		-dbPass $(IOT_DATABASE_PASS) \
		-dbUser $(IOT_DATABASE_USER) \
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
		-sonosRefreshToken "$(SONOS_REFRESH_TOKEN)" \
		-enedisEmail $(ENEDIS_EMAIL) \
		-enedisPassword $(ENEDIS_PASSWORD) \

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
