default: go docker

go: deps dev

dev: format lint tst bench build

docker: docker-deps docker-build

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/goimports
	dep ensure

format:
	goimports -w **/*.go *.go
	gofmt -s -w **/*.go *.go

lint:
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/iot iot.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -installsuffix nocgo -o bin/iot-arm iot.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -installsuffix nocgo -o bin/iot-arm64 iot.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/worker worker/worker.go

docker-deps:
	curl -s -o cacert.pem https://curl.haxx.se/ca/cacert.pem

docker-build:
	docker build -t ${DOCKER_USER}/iot .
	docker build -t ${DOCKER_USER}/iot:arm -f Dockerfile_arm .
	docker build -t ${DOCKER_USER}/iot:arm64 -f Dockerfile_arm64 .

docker-push:
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}
	docker push ${DOCKER_USER}/iot
	docker push ${DOCKER_USER}/iot:arm
	docker push ${DOCKER_USER}/iot:arm64

start-deps:
	go get -u github.com/ViBiOh/auth/bcrypt

start-api:
	go run iot.go \
		-tls=false \
		-authUsers admin:admin \
		-basicUsers "1:admin:`bcrypt admin`" \
		-secretKey SECRET_KEY \
		-csp "default-src 'self'; style-src 'self' 'unsafe-inline'"

start-worker:
	go run worker/worker.go \
		-websocket ws://localhost:1080/ws/hue \
		-secretKey SECRET_KEY \
		-hueConfig ./hue.json \
		-hueUsername ${BRIDGE_USERNAME} \
		-hueBridgeIP ${BRIDGE_IP} \
		-hueClean
