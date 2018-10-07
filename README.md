# iot

[![Build Status](https://travis-ci.org/ViBiOh/iot.svg?branch=master)](https://travis-ci.org/ViBiOh/iot)
[![codecov](https://codecov.io/gh/ViBiOh/iot/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/iot)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/iot)](https://goreportcard.com/report/github.com/ViBiOh/iot)

## Usage of Web server

```bash
Usage of iot:
  -authUrl string
      [auth] Auth URL, if remote
  -authUsers string
      [auth] List of allowed users and profiles (e.g. user:profile1|profile2,user2:profile3)
  -basicUsers string
      [Basic] Users in the form "id:username:password,id2:username2:password2"
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
  -dysonCountry string
      Dyson Link Country (default "FR")
  -dysonEmail string
      Dyson Link Email
  -dysonPassword string
      Dyson Link Password
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -port int
      Listen port (default 1080)
  -rollbarEnv string
      [rollbar] Environment (default "prod")
  -rollbarServerRoot string
      [rollbar] Server Root
  -rollbarToken string
      [rollbar] Token
  -secretKey string
      [iot] Secret Key between worker and API
  -sonosAccessToken string
      [sonos] Access Token
  -sonosClientID string
      [sonos] Client ID
  -sonosClientSecret string
      [sonos] Client Secret
  -sonosRefreshToken string
      [sonos] Refresh Token
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
      [health] User-Agent used (default "Golang alcotest")
```

## Usage of IoT worker

```bash
Usage of worker:
  -hueBridgeIP string
      [hue] IP of Bridge
  -hueClean
      [hue] Clean Hue
  -hueConfig string
      [hue] Configuration filename
  -hueUsername string
      [hue] Username for Bridge
  -netatmoAccessToken string
      [netatmo] Access Token
  -netatmoClientID string
      [netatmo] Client ID
  -netatmoClientSecret string
      [netatmo] Client Secret
  -netatmoRefreshToken string
      [netatmo] Refresh Token
  -rollbarEnv string
      [rollbar] Environment (default "prod")
  -rollbarServerRoot string
      [rollbar] Server Root
  -rollbarToken string
      [rollbar] Token
  -secretKey string
      Secret Key
  -tracingAgent string
      [opentracing] Jaeger Agent (e.g. host:port) (default "jaeger:6831")
  -tracingName string
      [opentracing] Service name
  -websocket string
      WebSocket URL
```

## Create systemd service for worker

Compile go binary

```bash
go install github.com/ViBiOh/iot/cmd/worker
```

Get username for Hue API by browsing `http://192.168.1.10/debug/clip.html`

```
POST /api
Body: {"devicetype":"iot-worker"}
```

Create file `sudo vi /lib/systemd/system/iot-worker.service`

```
[Unit]
Description=iot-worker
After=network.target

[Service]
Type=simple
User=vibioh
EnvironmentFile=/home/vibioh/.env
ExecStart=/home/vibioh/code/bin/worker -secretKey ${IOT_SECRET_KEY} -websocket wss://iot.vibioh.fr/ws -hueBridgeIP ${BRIDGE_IP} -hueUsername ${BRIDGE_USERNAME} -hueConfig /home/vibioh/code/src/github.com/ViBiOh/iot/hue.json -hueClean -netatmoAccessToken ${NETATMO_ACCESS_TOKEN} -netatmoClientID ${NETATMO_CLIENT_ID} -netatmoClientSecret ${NETATMO_CLIENT_SECRET} -netatmoRefreshToken ${NETATMO_REFRESH_TOKEN} -sonosAccessToken ${SONOS_ACCESS_TOKEN} -sonosClientID ${SONOS_CLIENT_ID} -sonosClientSecret ${SONOS_CLIENT_SECRET} -sonosRefreshToken ${SONOS_REFRESH_TOKEN} -tracingName iot_worker -tracingAgent vibioh.fr:6831
Restart=always
RestartSec=60s

[Install]
WantedBy=multi-user.target
```

Enable and start service

```bash
sudo systemctl daemon-reload
sudo systemctl enable iot-worker.service
sudo systemctl start iot-worker.service
journalctl -u iot-worker.service
```
