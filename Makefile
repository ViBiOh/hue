APP_NAME ?= iot
VERSION ?= $(shell git log --pretty=format:'%h' -n 1)
AUTHOR ?= $(shell git log --pretty=format:'%an' -n 1)

MAKEFLAGS += --silent
GOBIN=bin
BINARY_PATH=$(GOBIN)/$(APP_NAME)

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## $(APP_NAME): Build app with dependencies download
$(APP_NAME): deps go

.PHONY: go
go: format lint tst bench build

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
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck
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
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

## tst: Test code with coverage
.PHONY: tst
tst:
	script/coverage

## bench: Benchmark code
.PHONY: bench
bench:
	go test ./... -bench . -benchmem -run Benchmark.*

## build: Build binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH) cmd/api/iot.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH)-worker cmd/worker/worker.go

## start-deps: Download start dependencies
.PHONY: start-deps
start-deps:
	go get github.com/ViBiOh/auth/cmd/bcrypt

## start-worker: Start worker
.PHONY: start-worker
start-worker:
	go run cmd/worker/worker.go \
		-websocket ws://localhost:1080/ws/hue \
		-secretKey SECRET_KEY \
		-hueConfig ./hue.json \
		-hueUsername $(BRIDGE_USERNAME) \
		-hueBridgeIP $(BRIDGE_IP) \
		-hueClean

## start: Start app
.PHONY: start
start:
	go run cmd/api/iot.go \
		-tls=false \
		-authUsers admin:admin \
		-basicUsers "1:admin:`bcrypt admin`" \
		-secretKey SECRET_KEY \
		-csp "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'" \
		-netatmoAccessToken "$(NETATMO_ACCESS_TOKEN)" \
		-netatmoClientID "$(NETATMO_CLIENT_ID)" \
		-netatmoClientSecret "$(NETATMO_CLIENT_SECRET)" \
		-netatmoRefreshToken "$(NETATMO_REFRESH_TOKEN)" \
		-sonosAccessToken "$(SONOS_ACCESS_TOKEN)" \
		-sonosClientID "$(SONOS_CLIENT_ID)" \
		-sonosClientSecret "$(SONOS_CLIENT_SECRET)" \
		-sonosRefreshToken "$(SONOS_REFRESH_TOKEN)"
