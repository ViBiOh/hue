package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/alcotest/healthcheck"
	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
	"github.com/ViBiOh/iot/hue"
	"github.com/ViBiOh/iot/iot"
	"github.com/ViBiOh/iot/netatmo"
	"github.com/ViBiOh/iot/wemo"
)

const (
	websocketPath   = `/ws`
	healthcheckPath = `/health`
	wemoPath        = `/wemo`
	huePath         = `/hue`
)

var (
	apiHandler http.Handler
	iotHandler http.Handler

	hueHandler   http.Handler
	hueWsHandler http.Handler
	wemoHandler  http.Handler
)

var healthcheckHandler = http.StripPrefix(healthcheckPath, healthcheck.Handler())

func restHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, healthcheckPath) {
			healthcheckHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, wemoPath) {
			wemoHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, huePath) {
			hueHandler.ServeHTTP(w, r)
		} else {
			iotHandler.ServeHTTP(w, r)
		}
	})
}

func wsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, huePath) {
			hueWsHandler.ServeHTTP(w, r)
		} else {
			httputils.NotFound(w)
		}
	})
}

func handler() http.Handler {
	websocket := http.StripPrefix(websocketPath, wsHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, websocketPath) {
			websocket.ServeHTTP(w, r)
		} else {
			apiHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, true, `Serve TLS content`)
	alcotestConfig := alcotest.Flags(``)
	authConfig := auth.Flags(`auth`)
	certConfig := cert.Flags(`tls`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	netatmoConfig := netatmo.Flags(`netatmo`)
	wemoConfig := wemo.Flags(`wemo`)
	hueConfig := hue.Flags(`hue`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	log.Printf(`Starting server on port %s`, *port)

	netatmoApp := netatmo.NewApp(netatmoConfig)
	iotApp := iot.NewApp(authConfig, netatmoApp)
	wemoApp := wemo.NewApp(wemoConfig, iotApp)
	hueApp := hue.NewApp(hueConfig, iotApp)

	hueHandler = http.StripPrefix(huePath, hueApp.Handler())
	hueWsHandler = http.StripPrefix(huePath, hueApp.WebsocketHandler())
	wemoHandler = http.StripPrefix(wemoPath, wemoApp.Handler())
	iotHandler = gziphandler.GzipHandler(iotApp.Handler())

	apiHandler = prometheus.Handler(prometheusConfig, rate.Handler(rateConfig, owasp.Handler(owaspConfig, cors.Handler(corsConfig, restHandler()))))
	server := &http.Server{
		Addr:    `:` + *port,
		Handler: handler(),
	}

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *tls {
			log.Print(`Listening with TLS enabled`)
			serveError <- cert.ListenAndServeTLS(certConfig, server)
		} else {
			serveError <- server.ListenAndServe()
		}
	}()

	httputils.ServerGracefulClose(server, serveError, nil)
}
