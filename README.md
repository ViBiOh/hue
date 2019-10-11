# iot

[![Build Status](https://travis-ci.org/ViBiOh/iot.svg?branch=master)](https://travis-ci.org/ViBiOh/iot)
[![codecov](https://codecov.io/gh/ViBiOh/iot/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/iot)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/iot)](https://goreportcard.com/report/github.com/ViBiOh/iot)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=ViBiOh/iot)](https://dependabot.com)

## Usage of Web server

```bash
Usage of iot:
  -address string
        [http] Listen address {IOT_ADDRESS}
  -assetsDirectory string
        [iot] Assets directory (static and templates) {IOT_ASSETS_DIRECTORY}
  -cert string
        [http] Certificate file {IOT_CERT}
  -corsCredentials
        [cors] Access-Control-Allow-Credentials {IOT_CORS_CREDENTIALS}
  -corsExpose string
        [cors] Access-Control-Expose-Headers {IOT_CORS_EXPOSE}
  -corsHeaders string
        [cors] Access-Control-Allow-Headers {IOT_CORS_HEADERS} (default "Content-Type")
  -corsMethods string
        [cors] Access-Control-Allow-Methods {IOT_CORS_METHODS} (default "GET")
  -corsOrigin string
        [cors] Access-Control-Allow-Origin {IOT_CORS_ORIGIN} (default "*")
  -csp string
        [owasp] Content-Security-Policy {IOT_CSP} (default "default-src 'self'; base-uri 'self'")
  -frameOptions string
        [owasp] X-Frame-Options {IOT_FRAME_OPTIONS} (default "deny")
  -hsts
        [owasp] Indicate Strict Transport Security {IOT_HSTS} (default true)
  -key string
        [http] Key file {IOT_KEY}
  -mqttClientID string
        [mqtt] Client ID {IOT_MQTT_CLIENT_ID} (default "iot")
  -mqttPass string
        [mqtt] Password {IOT_MQTT_PASS}
  -mqttPort int
        [mqtt] Port {IOT_MQTT_PORT} (default 80)
  -mqttServer string
        [mqtt] Server name {IOT_MQTT_SERVER}
  -mqttUseTLS
        [mqtt] Use TLS {IOT_MQTT_USE_TLS} (default true)
  -mqttUser string
        [mqtt] Username {IOT_MQTT_USER}
  -port int
        [http] Listen port {IOT_PORT} (default 1080)
  -prometheus
        [iot] Expose Prometheus metrics {IOT_PROMETHEUS}
  -prometheusPath string
        [prometheus] Path for exposing metrics {IOT_PROMETHEUS_PATH} (default "/metrics")
  -publish string
        [iot] Topic to publish to {IOT_PUBLISH} (default "worker")
  -subscribe string
        [iot] Topic to subscribe to {IOT_SUBSCRIBE}
  -tracingAgent string
        [tracing] Jaeger Agent (e.g. host:port) {IOT_TRACING_AGENT} (default "jaeger:6831")
  -tracingName string
        [tracing] Service name {IOT_TRACING_NAME}
  -url string
        [alcotest] URL to check {IOT_URL}
  -userAgent string
        [alcotest] User-Agent for check {IOT_USER_AGENT} (default "Golang alcotest")
```

## Usage of IoT worker

```bash
Usage of iot-worker:
  -hueBridgeIP string
        [hue] IP of Bridge {IOT_WORKER_HUE_BRIDGE_IP}
  -hueConfig string
        [hue] Configuration filename {IOT_WORKER_HUE_CONFIG}
  -hueUsername string
        [hue] Username for Bridge {IOT_WORKER_HUE_USERNAME}
  -mqttClientID string
        [mqtt] Client ID {IOT_WORKER_MQTT_CLIENT_ID} (default "iot")
  -mqttPass string
        [mqtt] Password {IOT_WORKER_MQTT_PASS}
  -mqttPort int
        [mqtt] Port {IOT_WORKER_MQTT_PORT} (default 80)
  -mqttServer string
        [mqtt] Server name {IOT_WORKER_MQTT_SERVER}
  -mqttUseTLS
        [mqtt] Use TLS {IOT_WORKER_MQTT_USE_TLS} (default true)
  -mqttUser string
        [mqtt] Username {IOT_WORKER_MQTT_USER}
  -netatmoAccessToken string
        [netatmo] Access Token {IOT_WORKER_NETATMO_ACCESS_TOKEN}
  -netatmoClientID string
        [netatmo] Client ID {IOT_WORKER_NETATMO_CLIENT_ID}
  -netatmoClientSecret string
        [netatmo] Client Secret {IOT_WORKER_NETATMO_CLIENT_SECRET}
  -netatmoRefreshToken string
        [netatmo] Refresh Token {IOT_WORKER_NETATMO_REFRESH_TOKEN}
  -publish string
        [worker] Topics to publish to, comma separated {IOT_WORKER_PUBLISH} (default "local,remote")
  -sonosAccessToken string
        [sonos] Access Token {IOT_WORKER_SONOS_ACCESS_TOKEN}
  -sonosClientID string
        [sonos] Client ID {IOT_WORKER_SONOS_CLIENT_ID}
  -sonosClientSecret string
        [sonos] Client Secret {IOT_WORKER_SONOS_CLIENT_SECRET}
  -sonosRefreshToken string
        [sonos] Refresh Token {IOT_WORKER_SONOS_REFRESH_TOKEN}
  -subscribe string
        [worker] Topic to subscribe to {IOT_WORKER_SUBSCRIBE} (default "worker")
```

## Create systemd service for worker

Compile go binary

```bash
go install github.com/ViBiOh/iot/cmd/worker
go install github.com/ViBiOh/iot/cmd/iot
```

Get username for Hue API by browsing `http://192.168.1.10/debug/clip.html` and add credentials to `.env` file.

```
POST /api
Body: {"devicetype":"iot-worker"}
```

Enable and start services

```bash
make systemd
journalctl -u iot-remote-worker.service
```
