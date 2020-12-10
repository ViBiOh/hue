package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/hue/pkg/hue"
	"github.com/ViBiOh/hue/pkg/renderer"
)

const (
	apiPath = "/api"
)

func main() {
	fs := flag.NewFlagSet("hue", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'"))
	corsConfig := cors.Flags(fs, "cors")
	rendererConfig := renderer.Flags(fs, "", flags.NewOverride("Title", "Hue"), flags.NewOverride("PublicURL", "https://hue.vibioh.fr"))

	hueConfig := hue.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	prometheusApp := prometheus.New(prometheusConfig)
	prometheusRegisterer := prometheusApp.Registerer()

	rendererApp, err := renderer.New(rendererConfig, hue.FuncMap)
	logger.Fatal(err)

	hueApp, err := hue.New(hueConfig, prometheusRegisterer, rendererApp)
	logger.Fatal(err)

	rendererHandler := rendererApp.Handler(hueApp.TemplateFunc)
	hueHandler := http.StripPrefix(apiPath, hueApp.Handler())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, apiPath) {
			hueHandler.ServeHTTP(w, r)
			return
		}

		rendererHandler.ServeHTTP(w, r)
	})

	go hueApp.Start()

	httputils.New(serverConfig).ListenAndServe(handler, nil, prometheusApp.Middleware, owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware)
}
