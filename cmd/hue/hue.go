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
	"github.com/ViBiOh/hue/pkg/hue"
	"github.com/ViBiOh/hue/pkg/renderer"
)

const (
	faviconPath = "/favicon"
)

func main() {
	fs := flag.NewFlagSet("hue", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	hueConfig := hue.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	prometheusApp := prometheus.New(prometheusConfig)
	prometheusRegisterer := prometheusApp.Registerer()

	hueApp, err := hue.New(hueConfig, prometheusRegisterer)
	logger.Fatal(err)

	rendererApp, err := renderer.New(hueApp)
	logger.Fatal(err)

	rendererHandler := rendererApp.Handler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, path.Join("static", r.URL.Path))
		} else {
			rendererHandler.ServeHTTP(w, r)
		}
	})

	go hueApp.Start()

	server := httputils.New(serverConfig)
	server.Middleware(prometheusApp.Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)
	server.ListenServeWait(handler)
}
