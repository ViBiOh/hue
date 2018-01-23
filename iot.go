package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/alcotest/healthcheck"
	"github.com/ViBiOh/auth/auth"
	authProvider "github.com/ViBiOh/auth/provider"
	"github.com/ViBiOh/auth/provider/basic"
	authService "github.com/ViBiOh/auth/service"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
	"github.com/ViBiOh/iot/hue"
	"github.com/ViBiOh/iot/iot"
	"github.com/ViBiOh/iot/netatmo"
	"github.com/ViBiOh/iot/provider"
)

const (
	websocketPath   = `/ws`
	healthcheckPath = `/health`
	faviconPath     = `/favicon`
	huePath         = `/hue`

	webDirectory = `./web`
)

var (
	apiHandler http.Handler
	wsHandler  http.Handler

	iotHandler http.Handler
	hueHandler http.Handler
)

func restHandler(authApp *auth.App) http.Handler {
	healthcheckHandler := http.StripPrefix(healthcheckPath, healthcheck.Handler())

	authHandler := authApp.HandlerWithFail(func(w http.ResponseWriter, r *http.Request, _ *authProvider.User) {
		if strings.HasPrefix(r.URL.Path, huePath) {
			hueHandler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, webDirectory+r.URL.Path)
		} else {
			iotHandler.ServeHTTP(w, r)
		}
	}, func(w http.ResponseWriter, r *http.Request, err error) {
		if auth.IsForbiddenErr(err) {
			httputils.Forbidden(w)
		} else if err == auth.ErrEmptyAuthorization && authApp.URL != `` {
			http.Redirect(w, r, path.Join(authApp.URL, `/redirect/github`), http.StatusFound)
		} else {
			w.Header().Add(`WWW-Authenticate`, `Basic`)
			httputils.Unauthorized(w, err)
		}
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, healthcheckPath) {
			healthcheckHandler.ServeHTTP(w, r)
		} else {
			authHandler.ServeHTTP(w, r)
		}
	})
}

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, websocketPath) {
			wsHandler.ServeHTTP(w, r)
		} else {
			apiHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	port := flag.Int(`port`, 1080, `Listen port`)
	tls := flag.Bool(`tls`, true, `Serve TLS content`)
	alcotestConfig := alcotest.Flags(``)
	certConfig := cert.Flags(`tls`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	authConfig := auth.Flags(`auth`)
	authBasicConfig := basic.Flags(`basic`)

	iotConfig := iot.Flags(``)
	netatmoConfig := netatmo.Flags(`netatmo`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	log.Printf(`Starting server on port %d`, *port)

	authApp := auth.NewApp(authConfig, authService.NewBasicApp(authBasicConfig))
	netatmoApp := netatmo.NewApp(netatmoConfig)
	hueApp := hue.NewApp()
	iotApp := iot.NewApp(iotConfig, map[string]provider.Provider{
		`Netatmo`: netatmoApp,
		`Hue`:     hueApp,
	})

	hueHandler = http.StripPrefix(huePath, hueApp.Handler())
	iotHandler = gziphandler.GzipHandler(iotApp.Handler())
	wsHandler = http.StripPrefix(websocketPath, iotApp.WebsocketHandler())

	apiHandler = prometheus.Handler(prometheusConfig, rate.Handler(rateConfig, owasp.Handler(owaspConfig, cors.Handler(corsConfig, restHandler(authApp)))))
	server := &http.Server{
		Addr:    fmt.Sprintf(`:%d`, *port),
		Handler: handler(),
	}

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *tls {
			log.Print(`Listening with TLS enabled`)
			serveError <- cert.ListenAndServeTLS(certConfig, server)
		} else {
			log.Print(`⚠ iot is running without secure connection ⚠`)
			serveError <- server.ListenAndServe()
		}
	}()

	httputils.ServerGracefulClose(server, serveError, nil)
}
