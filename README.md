# hue

[![Build](https://github.com/ViBiOh/hue/workflows/Build/badge.svg)](https://github.com/ViBiOh/hue/actions)

A web interface for easily managing your Hue installation.

![](preview.png)

## Getting started

Golang binary is built with static link. You can download it directly from the [GitHub Release page](https://github.com/ViBiOh/hue/releases) or build it by yourself by cloning this repo and running `make`.

You can configure app by passing CLI args or environment variables (cf. [Usage](#usage) section). The args override environment variables.

It's a single static binary with embedded templates and static. No Javascript framework. HTTP and HTML have all we need. The recommended way to use it is the Docker container but binary is self-sufficient too.

You'll find a Kubernetes exemple in the [`infra/`](infra) folder. It contains two ingresses : one for "same network access" and another, publicly available but with basic-auth.

### Get credentials from bridge

To connect to your bridge, you'll need credentials generated by Hue Bridge.

Get username for Hue API by browsing `http://192.168.1.10/debug/clip.html` and retrieve username credentials.

```
POST /api
Body: {"devicetype":"hue"}
```

### Using it

It's recommended to use the official Hue mobile app for setupping and configuring your devices. The goal of this project is to provide an easy-to-use web interface for controlling the lights.

When turning lights on, there are 3 availables states:

- `on`: 100% brightness in 5 seconds, to avoid eyes bleeding
- `half`: 50% brightness in 5 seconds, for some ambiance lighting
- `dimmed`: minimum brightness in 5 seconds, for very low light need

You can use this software to configure a subset of your Hue installation:

- Hue Tap buttons behaviors
- Hue Motion Sensor behaviors
- Schedule light on/off based on time

It also supports some third-party devices that are compatible with the Hub, such a power-switch. In this case there is only two mode : on/off.

### Why ?

Most IoT devices and platforms are relying on applications installed on your smartphone. But if you're not alone at home, you have to share your credentials with others, which is a wrong security pattern.

This interface allows anybody who have access to manage lights. In a local network, with a properly configured firewall, anyone on the Wi-Fi can turn off/on lights, from mobile-ready dark mode web interface.

The application is not an app or hub replacement, it uses the Hub's API and is not required to continue to use your existing remotes ou app like you already do.

### Metrics

The web service exposes multiples metrics gathered from the motions sensors and taps: the battery life, the temperature and the motion detection. They are available with OpenTelemetry.

## Usage

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of hue:
  --address           string    [server] Listen address ${HUE_ADDRESS}
  --bridgeIP          string    [hue] IP of Bridge ${HUE_BRIDGE_IP}
  --cert              string    [server] Certificate file ${HUE_CERT}
  --config            string    [hue] Configuration filename ${HUE_CONFIG}
  --corsCredentials             [cors] Access-Control-Allow-Credentials ${HUE_CORS_CREDENTIALS} (default false)
  --corsExpose        string    [cors] Access-Control-Expose-Headers ${HUE_CORS_EXPOSE}
  --corsHeaders       string    [cors] Access-Control-Allow-Headers ${HUE_CORS_HEADERS} (default "Content-Type")
  --corsMethods       string    [cors] Access-Control-Allow-Methods ${HUE_CORS_METHODS} (default "GET")
  --corsOrigin        string    [cors] Access-Control-Allow-Origin ${HUE_CORS_ORIGIN} (default "*")
  --csp               string    [owasp] Content-Security-Policy ${HUE_CSP} (default "default-src 'self'; script-src 'httputils-nonce'; style-src 'httputils-nonce'")
  --frameOptions      string    [owasp] X-Frame-Options ${HUE_FRAME_OPTIONS} (default "deny")
  --graceDuration     duration  [http] Grace duration when signal received ${HUE_GRACE_DURATION} (default 30s)
  --hsts                        [owasp] Indicate Strict Transport Security ${HUE_HSTS} (default true)
  --idleTimeout       duration  [server] Idle Timeout ${HUE_IDLE_TIMEOUT} (default 2m0s)
  --key               string    [server] Key file ${HUE_KEY}
  --loggerJson                  [logger] Log format as JSON ${HUE_LOGGER_JSON} (default false)
  --loggerLevel       string    [logger] Logger level ${HUE_LOGGER_LEVEL} (default "INFO")
  --loggerLevelKey    string    [logger] Key for level in JSON ${HUE_LOGGER_LEVEL_KEY} (default "level")
  --loggerMessageKey  string    [logger] Key for message in JSON ${HUE_LOGGER_MESSAGE_KEY} (default "msg")
  --loggerTimeKey     string    [logger] Key for timestamp in JSON ${HUE_LOGGER_TIME_KEY} (default "time")
  --minify                      Minify HTML ${HUE_MINIFY} (default true)
  --name              string    [server] Name ${HUE_NAME} (default "http")
  --okStatus          int       [http] Healthy HTTP Status code ${HUE_OK_STATUS} (default 204)
  --pathPrefix        string    Root Path Prefix ${HUE_PATH_PREFIX}
  --port              uint      [server] Listen port (0 to disable) ${HUE_PORT} (default 1080)
  --pprofAgent        string    [pprof] URL of the Datadog Trace Agent (e.g. http://datadog.observability:8126) ${HUE_PPROF_AGENT}
  --pprofPort         int       [pprof] Port of the HTTP server (0 to disable) ${HUE_PPROF_PORT} (default 0)
  --publicURL         string    Public URL ${HUE_PUBLIC_URL} (default "https://hue.vibioh.fr")
  --readTimeout       duration  [server] Read Timeout ${HUE_READ_TIMEOUT} (default 5s)
  --shutdownTimeout   duration  [server] Shutdown Timeout ${HUE_SHUTDOWN_TIMEOUT} (default 10s)
  --telemetryRate     string    [telemetry] OpenTelemetry sample rate, 'always', 'never' or a float value ${HUE_TELEMETRY_RATE} (default "always")
  --telemetryURL      string    [telemetry] OpenTelemetry gRPC endpoint (e.g. otel-exporter:4317) ${HUE_TELEMETRY_URL}
  --telemetryUint64             [telemetry] Change OpenTelemetry Trace ID format to an unsigned int 64 ${HUE_TELEMETRY_UINT64} (default true)
  --title             string    Application title ${HUE_TITLE} (default "Hue")
  --update                      [hue] Update configuration from file ${HUE_UPDATE} (default false)
  --url               string    [alcotest] URL to check ${HUE_URL}
  --userAgent         string    [alcotest] User-Agent for check ${HUE_USER_AGENT} (default "Alcotest")
  --username          string    [hue] Username for Bridge ${HUE_USERNAME}
  --v2BridgeIP        string    [v2] IP of Bridge ${HUE_V2_BRIDGE_IP}
  --v2Config          string    [v2] Configuration filename ${HUE_V2_CONFIG}
  --v2Username        string    [v2] Username for Bridge ${HUE_V2_USERNAME}
  --writeTimeout      duration  [server] Write Timeout ${HUE_WRITE_TIMEOUT} (default 10s)
```
