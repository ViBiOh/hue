# iot

[![Build Status](https://travis-ci.com/ViBiOh/iot.svg?branch=master)](https://travis-ci.com/ViBiOh/iot)
[![codecov](https://codecov.io/gh/ViBiOh/iot/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/iot)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/iot)](https://goreportcard.com/report/github.com/ViBiOh/iot)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=ViBiOh/iot)](https://dependabot.com)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ViBiOh_iot&metric=alert_status)](https://sonarcloud.io/dashboard?id=ViBiOh_iot)

## Usage of Web server

```bash
Usage of hue:
  -address string
        [http] Listen address {HUE_ADDRESS}
  -bridgeIP string
        [hue] IP of Bridge {HUE_BRIDGE_IP}
  -cert string
        [http] Certificate file {HUE_CERT}
  -config string
        [hue] Configuration filename {HUE_CONFIG}
  -corsCredentials
        [cors] Access-Control-Allow-Credentials {HUE_CORS_CREDENTIALS}
  -corsExpose string
        [cors] Access-Control-Expose-Headers {HUE_CORS_EXPOSE}
  -corsHeaders string
        [cors] Access-Control-Allow-Headers {HUE_CORS_HEADERS} (default "Content-Type")
  -corsMethods string
        [cors] Access-Control-Allow-Methods {HUE_CORS_METHODS} (default "GET")
  -corsOrigin string
        [cors] Access-Control-Allow-Origin {HUE_CORS_ORIGIN} (default "*")
  -csp string
        [owasp] Content-Security-Policy {HUE_CSP} (default "default-src 'self'; base-uri 'self'")
  -frameOptions string
        [owasp] X-Frame-Options {HUE_FRAME_OPTIONS} (default "deny")
  -graceDuration string
        [http] Grace duration when SIGTERM received {HUE_GRACE_DURATION} (default "15s")
  -hsts
        [owasp] Indicate Strict Transport Security {HUE_HSTS} (default true)
  -key string
        [http] Key file {HUE_KEY}
  -okStatus int
        [http] Healthy HTTP Status code {HUE_OK_STATUS} (default 204)
  -port uint
        [http] Listen port {HUE_PORT} (default 1080)
  -prometheusPath string
        [prometheus] Path for exposing metrics {HUE_PROMETHEUS_PATH} (default "/metrics")
  -url string
        [alcotest] URL to check {HUE_URL}
  -userAgent string
        [alcotest] User-Agent for check {HUE_USER_AGENT} (default "Alcotest")
  -username string
        [hue] Username for Bridge {HUE_USERNAME}
```

## Get credentials from bridge

Get username for Hue API by browsing `http://192.168.1.10/debug/clip.html` and add credentials to `.env` file.

```
POST /api
Body: {"devicetype":"hue"}
```

