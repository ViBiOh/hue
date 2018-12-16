package main

import (
	"flag"
	"net/http"
	"path"
	"strings"

	"github.com/ViBiOh/auth/pkg/auth"
	"github.com/ViBiOh/auth/pkg/ident/basic"
	authService "github.com/ViBiOh/auth/pkg/ident/service"
	httputils "github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/gzip"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/httputils/pkg/prometheus"
	"github.com/ViBiOh/httputils/pkg/server"
	"github.com/ViBiOh/iot/pkg/dyson"
	"github.com/ViBiOh/iot/pkg/hue"
	"github.com/ViBiOh/iot/pkg/iot"
	"github.com/ViBiOh/iot/pkg/netatmo"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/ViBiOh/iot/pkg/sonos"
)

const (
	websocketPath   = `/ws`
	healthcheckPath = `/health`
	faviconPath     = `/favicon`
	huePath         = `/hue`
	dysonPath       = `/dyson`
	sonosPath       = `/sonos`
)

func main() {
	serverConfig := httputils.Flags(``)
	alcotestConfig := alcotest.Flags(``)
	prometheusConfig := prometheus.Flags(`prometheus`)
	opentracingConfig := opentracing.Flags(`tracing`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	authConfig := auth.Flags(`auth`)
	authBasicConfig := basic.Flags(`basic`)
	iotConfig := iot.Flags(``)
	netatmoConfig := netatmo.Flags(`netatmo`)
	dysonConfig := dyson.Flags(`dyson`)
	sonosConfig := sonos.Flags(`sonos`)

	assetsDirectory := flag.String(`assetsDirectory`, `./`, `Assets directory (static and templates)`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	serverApp := httputils.NewApp(serverConfig)
	healthcheckApp := healthcheck.NewApp()
	prometheusApp := prometheus.NewApp(prometheusConfig)
	opentracingApp := opentracing.NewApp(opentracingConfig)
	gzipApp := gzip.NewApp()
	owaspApp := owasp.NewApp(owaspConfig)
	corsApp := cors.NewApp(corsConfig)

	authApp := auth.NewServiceApp(authConfig, authService.NewBasicApp(authBasicConfig, nil))
	netatmoApp := netatmo.NewApp(netatmoConfig)
	dysonApp := dyson.NewApp(dysonConfig)
	sonosApp := sonos.NewApp(sonosConfig)
	hueApp := hue.NewApp()
	iotApp := iot.NewApp(iotConfig, *assetsDirectory, map[string]provider.Provider{
		`Netatmo`: netatmoApp,
		`Hue`:     hueApp,
		`Dyson`:   dysonApp,
		`Sonos`:   sonosApp,
	})

	hueHandler := http.StripPrefix(huePath, hueApp.Handler())
	dysonHandler := http.StripPrefix(dysonPath, dysonApp.Handler())
	sonosHandler := http.StripPrefix(sonosPath, sonosApp.Handler())
	iotHandler := iotApp.Handler()
	wsHandler := http.StripPrefix(websocketPath, iotApp.WebsocketHandler())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, huePath) {
			hueHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, dysonPath) {
			dysonHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, sonosPath) {
			sonosHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, path.Join(*assetsDirectory, `static`, r.URL.Path))
		} else {
			iotHandler.ServeHTTP(w, r)
		}
	})

	apiHandler := server.ChainMiddlewares(handler, prometheusApp, opentracingApp, gzipApp, owaspApp, corsApp, authApp)

	serverApp.ListenAndServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, websocketPath) {
			wsHandler.ServeHTTP(w, r)
		} else {
			apiHandler.ServeHTTP(w, r)
		}
	}), nil, healthcheckApp)
}
