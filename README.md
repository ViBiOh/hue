# iot

[![Build Status](https://travis-ci.org/ViBiOh/iot.svg?branch=master)](https://travis-ci.org/ViBiOh/iot)
[![codecov](https://codecov.io/gh/ViBiOh/iot/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/iot)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/iot)](https://goreportcard.com/report/github.com/ViBiOh/iot)

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
  -hsts
      [owasp] Indicate Strict Transport Security (default true)
  -iftttWebHook string
      IFTTT WebHook Key
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
```