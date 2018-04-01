package main

import (
	"net/http"
	"path"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/auth/pkg/auth"
	authProvider "github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/auth/pkg/provider/basic"
	authService "github.com/ViBiOh/auth/pkg/service"
	"github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/iot/pkg/hue"
	"github.com/ViBiOh/iot/pkg/iot"
	"github.com/ViBiOh/iot/pkg/netatmo"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	websocketPath   = `/ws`
	healthcheckPath = `/health`
	faviconPath     = `/favicon`
	huePath         = `/hue`

	webDirectory = `./templates`
)

func main() {
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	authConfig := auth.Flags(`auth`)
	authBasicConfig := basic.Flags(`basic`)
	iotConfig := iot.Flags(``)
	netatmoConfig := netatmo.Flags(`netatmo`)

	httputils.NewApp(httputils.Flags(``), func() http.Handler {
		authApp := auth.NewApp(authConfig, authService.NewBasicApp(authBasicConfig))
		netatmoApp := netatmo.NewApp(netatmoConfig)
		hueApp := hue.NewApp()
		iotApp := iot.NewApp(iotConfig, map[string]provider.Provider{
			`Netatmo`: netatmoApp,
			`Hue`:     hueApp,
		})

		hueHandler := http.StripPrefix(huePath, hueApp.Handler())
		iotHandler := gziphandler.GzipHandler(iotApp.Handler())
		wsHandler := http.StripPrefix(websocketPath, iotApp.WebsocketHandler())

		healthcheckHandler := http.StripPrefix(healthcheckPath, healthcheck.Handler())

		authHandler := authApp.HandlerWithFail(func(w http.ResponseWriter, r *http.Request, _ *authProvider.User) {
			if strings.HasPrefix(r.URL.Path, huePath) {
				hueHandler.ServeHTTP(w, r)
			} else if strings.HasPrefix(r.URL.Path, faviconPath) {
				http.ServeFile(w, r, path.Join(webDirectory, r.URL.Path))
			} else {
				iotHandler.ServeHTTP(w, r)
			}
		}, func(w http.ResponseWriter, r *http.Request, err error) {
			if auth.IsForbiddenErr(err) {
				httperror.Forbidden(w)
			} else if err == auth.ErrEmptyAuthorization && authApp.URL != `` {
				http.Redirect(w, r, path.Join(authApp.URL, `/redirect/github`), http.StatusFound)
			} else {
				w.Header().Add(`WWW-Authenticate`, `Basic`)
				httperror.Unauthorized(w, err)
			}
		})

		restHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, healthcheckPath) {
				healthcheckHandler.ServeHTTP(w, r)
			} else {
				authHandler.ServeHTTP(w, r)
			}
		})

		apiHandler := owasp.Handler(owaspConfig, cors.Handler(corsConfig, restHandler))

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, websocketPath) {
				wsHandler.ServeHTTP(w, r)
			} else {
				apiHandler.ServeHTTP(w, r)
			}
		})
	}, nil).ListenAndServe()
}
