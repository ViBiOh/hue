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

## Usage

```
Usage of iot:
  -authUrl string
      [auth] Auth URL
  -authUsers string
      [auth] List of allowed users and profiles (e.g. user:profile1|profile2,user2:profile3)
  -c string
      [health] URL to check
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
      [owasp] Content-Security-Policy (default "default-src 'self'")
  -frameOptions string
      [owasp] X-Frame-Options (default "deny")
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -hueSecretKey string
      [hue] Secret Key between worker and API
  -netatmoAccessToken string
      [netatmo] Access Token
  -netatmoClientID string
      [netatmo] Client ID
  -netatmoClientSecret string
      [netatmo] Client Secret
  -netatmoRefreshToken string
      [netatmo] Refresh Token
  -port string
      Listen port (default "1080")
  -prometheusMetricsHost string
      [prometheus] Allowed hostname to call metrics endpoint (default "localhost")
  -prometheusMetricsPath string
      [prometheus] Metrics endpoint path (default "/metrics")
  -prometheusPrefix string
      [prometheus] Prefix (default "http")
  -rateCount uint
      [rate] IP limit (default 5000)
  -tls
      Serve TLS content (default true)
  -tlsCert string
      [tls] PEM Certificate file
  -tlsHosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -tlsKey string
      [tls] PEM Key file
  -wemoWebHook string
      [wemo] WebHook Key
```