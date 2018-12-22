package main

import (
	"flag"
	"net/http"
	"os"
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
	"github.com/ViBiOh/httputils/pkg/logger"
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
	fs := flag.NewFlagSet(`iot`, flag.ExitOnError)

	serverConfig := httputils.Flags(fs, ``)
	alcotestConfig := alcotest.Flags(fs, ``)
	prometheusConfig := prometheus.Flags(fs, `prometheus`)
	opentracingConfig := opentracing.Flags(fs, `tracing`)
	owaspConfig := owasp.Flags(fs, ``)
	corsConfig := cors.Flags(fs, `cors`)

	authConfig := auth.Flags(fs, `auth`)
	authBasicConfig := basic.Flags(fs, `basic`)
	iotConfig := iot.Flags(fs, ``)
	dysonConfig := dyson.Flags(fs, `dyson`)

	assetsDirectory := fs.String(`assetsDirectory`, `./`, `Assets directory (static and templates)`)

	if err := fs.Parse(os.Args[1:]); err != nil {
		logger.Fatal(`%+v`, err)
	}

	alcotest.DoAndExit(alcotestConfig)

	serverApp := httputils.New(serverConfig)
	healthcheckApp := healthcheck.New()
	prometheusApp := prometheus.New(prometheusConfig)
	opentracingApp := opentracing.New(opentracingConfig)
	gzipApp := gzip.New()
	owaspApp := owasp.New(owaspConfig)
	corsApp := cors.New(corsConfig)

	authApp := auth.NewService(authConfig, authService.NewBasic(authBasicConfig, nil))
	netatmoApp := netatmo.New()
	sonosApp := sonos.New()
	dysonApp := dyson.New(dysonConfig)
	hueApp := hue.New()
	iotApp := iot.New(iotConfig, *assetsDirectory, map[string]provider.Provider{
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
