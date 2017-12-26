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

const websocketPath = `/ws`
const healthcheckPath = `/health`
const wemoPath = `/wemo`
const huePath = `/hue`

var apiHandler http.Handler
var iotHandler http.Handler
var healthcheckHandler = http.StripPrefix(healthcheckPath, healthcheck.Handler())
var wemoHandler = http.StripPrefix(wemoPath, wemo.Handler())
var hueHandler = http.StripPrefix(huePath, hue.Handler())
var hueWsHandler = http.StripPrefix(huePath, hue.WebsocketHandler())

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

	if err := iot.Init(authConfig); err != nil {
		log.Printf(`Error while initializing iot: %v`, err)
	}
	if err := hue.Init(hueConfig); err != nil {
		log.Printf(`Error while initializing hue: %v`, err)
	}
	if err := wemo.Init(wemoConfig); err != nil {
		log.Printf(`Error while initializing wemo: %v`, err)
	}
	if err := netatmo.Init(netatmoConfig); err != nil {
		log.Printf(`Error while initializing netatmo: %v`, err)
	}

	log.Printf(`Starting server on port %s`, *port)

	iotHandler = gziphandler.GzipHandler(iot.Handler())
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
