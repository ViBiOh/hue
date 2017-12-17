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
	"github.com/ViBiOh/iot/iot"
)

const healthcheckPath = `/health`

var iotHandler http.Handler
var healthcheckHandler = http.StripPrefix(healthcheckPath, healthcheck.Handler())

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, healthcheckPath) {
			healthcheckHandler.ServeHTTP(w, r)
		} else {
			iotHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, true, `Serve TLS content`)
	iftttWebHook := flag.String(`iftttWebHook`, ``, `IFTTT WebHook Key`)
	alcotestConfig := alcotest.Flags(``)
	authConfig := auth.Flags(`auth`)
	certConfig := cert.Flags(`tls`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	if err := iot.Init(authConfig, *iftttWebHook); err != nil {
		log.Printf(`Error while initializing iot Handler: %v`, err)
	}
	iotHandler = gziphandler.GzipHandler(iot.Handler())

	log.Printf(`Starting server on port %s`, *port)

	server := &http.Server{
		Addr:    `:` + *port,
		Handler: prometheus.Handler(prometheusConfig, rate.Handler(rateConfig, owasp.Handler(owaspConfig, cors.Handler(corsConfig, handler())))),
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
