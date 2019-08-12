package main

import (
	"flag"
	"net/http"
	"os"
	"path"
	"strings"

	httputils "github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/httputils/pkg/prometheus"
	"github.com/ViBiOh/iot/pkg/hue"
	"github.com/ViBiOh/iot/pkg/iot"
	"github.com/ViBiOh/iot/pkg/mqtt"
	"github.com/ViBiOh/iot/pkg/netatmo"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/ViBiOh/iot/pkg/sonos"
)

const (
	healthcheckPath = "/health"
	faviconPath     = "/favicon"
	huePath         = "/hue"
	sonosPath       = "/sonos"
)

func main() {
	fs := flag.NewFlagSet("iot", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	opentracingConfig := opentracing.Flags(fs, "tracing")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	mqttConfig := mqtt.Flags(fs, "mqtt")
	iotConfig := iot.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	serverApp, err := httputils.New(serverConfig)
	logger.Fatal(err)

	prometheusApp := prometheus.New(prometheusConfig)
	opentracingApp := opentracing.New(opentracingConfig)
	owaspApp := owasp.New(owaspConfig)
	corsApp := cors.New(corsConfig)

	mqttApp, err := mqtt.New(mqttConfig)
	logger.Fatal(err)

	netatmoApp := netatmo.New()
	sonosApp := sonos.New()
	hueApp := hue.New()
	iotApp := iot.New(iotConfig, map[string]provider.Provider{
		"Netatmo": netatmoApp,
		"Hue":     hueApp,
		"Sonos":   sonosApp,
	}, mqttApp)

	hueHandler := http.StripPrefix(huePath, hueApp.Handler())
	sonosHandler := http.StripPrefix(sonosPath, sonosApp.Handler())
	iotHandler := iotApp.Handler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, huePath) {
			hueHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, sonosPath) {
			sonosHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, path.Join(*iotConfig.AssetsDirectory, "static", r.URL.Path))
		} else {
			iotHandler.ServeHTTP(w, r)
		}
	})

	iotApp.HandleWorker()

	serverApp.ListenAndServe(httputils.ChainMiddlewares(handler, prometheusApp, opentracingApp, owaspApp, corsApp), httputils.HealthHandler(nil), nil)
}
