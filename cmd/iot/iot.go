package main

import (
	"flag"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/iot/pkg/hue"
	"github.com/ViBiOh/iot/pkg/iot"
	"github.com/ViBiOh/iot/pkg/mqtt"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	healthcheckPath = "/health"
	faviconPath     = "/favicon"
	huePath         = "/hue"
)

func main() {
	fs := flag.NewFlagSet("iot", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	mqttConfig := mqtt.Flags(fs, "mqtt")
	iotConfig := iot.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	prometheusApp := prometheus.New(prometheusConfig)
	prometheusRegisterer := prometheusApp.Registerer()

	mqttApp, err := mqtt.New(mqttConfig)
	logger.Fatal(err)

	hueApp := hue.New(prometheusRegisterer)
	iotApp := iot.New(iotConfig, map[string]provider.Provider{
		"Hue": hueApp,
	}, mqttApp)

	hueHandler := http.StripPrefix(huePath, hueApp.Handler())
	iotHandler := iotApp.Handler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, huePath) {
			hueHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, path.Join(*iotConfig.AssetsDirectory, "static", r.URL.Path))
		} else {
			iotHandler.ServeHTTP(w, r)
		}
	})

	iotApp.HandleWorker()

	server := httputils.New(serverConfig)
	server.Middleware(prometheusApp.Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)
	server.ListenServeWait(handler)
}
