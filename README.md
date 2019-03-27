# iot

[![Build Status](https://travis-ci.org/ViBiOh/iot.svg?branch=master)](https://travis-ci.org/ViBiOh/iot)
[![codecov](https://codecov.io/gh/ViBiOh/iot/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/iot)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/iot)](https://goreportcard.com/report/github.com/ViBiOh/iot)

## Usage of Web server

```bash
Usage of iot:
  -assetsDirectory string
      [iot] Assets directory (static and templates)
  -corsCredentials
      [cors] Access-Control-Allow-Credentials
  -corsExpose string
      [cors] Access-Control-Expose-Headers
  -corsHeaders string
      [cors] Access-Control-Allow-Headers (default "Content-Type")
  -corsMethods string
      [cors] Access-Control-Allow-Methods (default "GET")
  -corsOrigin string
      [cors] Access-Control-Allow-Origin (default "*")
  -csp string
      [owasp] Content-Security-Policy (default "default-src 'self'; base-uri 'self'")
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -mqttClientID string
      [mqtt] Client ID (default "iot")
  -mqttPass string
      [mqtt] Password
  -mqttPort int
      [mqtt] Port (default 80)
  -mqttServer string
      [mqtt] Server name
  -mqttUseTLS
      [mqtt] Use TLS (default true)
  -mqttUser string
      [mqtt] Username
  -port int
      Listen port (default 1080)
  -prometheus
      [iot] Expose Prometheus metrics
  -prometheusPath string
      [prometheus] Path for exposing metrics (default "/metrics")
  -publish string
      [iot] Topic to publish to (default "worker")
  -subscribe string
      [iot] Topic to subscribe to
  -tls
      Serve TLS content (default true)
  -tlsCert string
      [tls] PEM Certificate file
  -tlsHosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -tlsKey string
      [tls] PEM Key file
  -tlsOrganization string
      [tls] Self-signed certificate organization (default "ViBiOh")
  -tracingAgent string
      [opentracing] Jaeger Agent (e.g. host:port) (default "jaeger:6831")
  -tracingName string
      [opentracing] Service name
  -url string
      [health] URL to check
  -userAgent string
      [health] User-Agent for check (default "Golang alcotest")
```

## Usage of IoT worker

```bash
Usage of iot-worker:
  -dysonClientID string
        [dyson] MQTT Client ID (default "iot")
  -dysonCountry string
        [dyson] Link eountry (default "FR")
  -dysonEmail string
        [dyson] Link email
  -dysonPassword string
        [dyson] Link eassword
  -enedisEmail string
        [enedis] Email
  -enedisPassword string
        [enedis] Password
  -hueBridgeIP string
        [hue] IP of Bridge
  -hueClean
        [hue] Clean Hue
  -hueConfig string
        [hue] Configuration filename
  -hueUsername string
        [hue] Username for Bridge
  -mqttClientID string
        [mqtt] Client ID (default "iot")
  -mqttPass string
        [mqtt] Password
  -mqttPort int
        [mqtt] Port (default 80)
  -mqttServer string
        [mqtt] Server name
  -mqttUseTLS
        [mqtt] Use TLS (default true)
  -mqttUser string
        [mqtt] Username
  -netatmoAccessToken string
        [netatmo] Access Token
  -netatmoClientID string
        [netatmo] Client ID
  -netatmoClientSecret string
        [netatmo] Client Secret
  -netatmoRefreshToken string
        [netatmo] Refresh Token
  -publish string
        Topics to publish to, comma separated (default "local,remote")
  -sonosAccessToken string
        [sonos] Access Token
  -sonosClientID string
        [sonos] Client ID
  -sonosClientSecret string
        [sonos] Client Secret
  -sonosRefreshToken string
        [sonos] Refresh Token
  -subscribe string
        Topic to subscribe to (default "worker")
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
