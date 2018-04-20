default: api docker-api

api: deps go

go: format lint tst bench build

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/goimports
	dep ensure

format:
	goimports -w */*/*.go
	gofmt -s -w */*/*.go

lint:
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build: build-api build-worker

build-api:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/iot cmd/api/iot.go

build-worker:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/worker cmd/worker/worker.go

docker-deps:
	curl -s -o cacert.pem https://curl.haxx.se/ca/cacert.pem

docker-login:
	echo $(DOCKER_PASS) | docker login -u $(DOCKER_USER) --password-stdin

docker-api: docker-build-api docker-push-api

docker-build-api: docker-deps
	docker build -t $(DOCKER_USER)/iot .

docker-push-api: docker-login
	docker push $(DOCKER_USER)/iot

start-deps:
	go get -u github.com/ViBiOh/auth/cmd/bcrypt

start-api:
	go run cmd/api/iot.go \
		-tls=false \
		-authUsers admin:admin \
		-basicUsers "1:admin:`bcrypt admin`" \
		-secretKey SECRET_KEY \
		-csp "default-src 'self'; style-src 'self' 'unsafe-inline'"

start-worker:
	go run cmd/worker/worker.go \
		-websocket ws://localhost:1080/ws/hue \
		-secretKey SECRET_KEY \
		-hueConfig ./hue.json \
		-hueUsername $(BRIDGE_USERNAME) \
		-hueBridgeIP $(BRIDGE_IP) \
		-hueClean \
		-debug

.PHONY: api go deps format lint tst bench build build-api build-worker docker-deps docker-api docker-login docker-build-api docker-push-api start-deps start-api start-worker
