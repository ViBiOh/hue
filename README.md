# iot

[![Build Status](https://travis-ci.org/ViBiOh/iot.svg?branch=master)](https://travis-ci.org/ViBiOh/iot)
[![codecov](https://codecov.io/gh/ViBiOh/iot/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/iot)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/iot)](https://goreportcard.com/report/github.com/ViBiOh/iot)

## Generate NetAtmo Token

```bash
export NETATMO_USER=[YOUR_EMAIL]
export NETATMO_PASS=[YOUR_PASS]
export NETATMO_CLIENT_ID=[YOUR_CLIENT_ID]
export NETATMO_CLIENT_SECRET=[YOUR_SECRET_ID]
export NETATMO_SCOPES=read_station
curl -X POST https://api.netatmo.com/oauth2/token --data "grant_type=password&username=${NETATMO_USER}&password=${NETATMO_PASS}&client_id=${NETATMO_CLIENT_ID}&client_secret=${NETATMO_CLIENT_SECRET}&scope=${NETATMO_SCOPES}"
```

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
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -netatmoAccessToken string
      [netatmo] Access Token
  -netatmoClientID string
      [netatmo] Client ID
  -netatmoClientSecret string
      [netatmo] Client Secret
  -netatmoRefreshToken string
      [netatmo] Refresh Token
  -port int
      Listen port (default 1080)
  -secretKey string
      [iot] Secret Key between worker and API
  -tls
      Serve TLS content (default true)
  -tlsCert string
      [tls] PEM Certificate file
  -tlsHosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -tlsKey string
      [tls] PEM Key file
  -url string
      [health] URL to check
```

## Usage of IoT worker

```bash
Usage of worker:
  -debug
      Enable debug
  -hueBridgeIP string
      [hue] IP of Bridge
  -hueClean
      [hue] Clean Hue
  -hueConfig string
      [hue] Configuration filename
  -hueUsername string
      [hue] Username for Bridge
  -secretKey string
      Secret Key
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
ExecStart=/home/vibioh/code/bin/worker -secretKey SECRET_KEY -websocket wss://iot.vibioh.fr -hueBridgeIP 192.168.1.10 -hueUsername HUE_USERNAME -hueConfig /home/vibioh/code/src/github.com/ViBiOh/iot/hue.json -hueClean
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
