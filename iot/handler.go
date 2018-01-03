package iot

import (
	"html/template"
	"net/http"
	"path"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/iot/provider"
)

// App stores informations and secret of API
type App struct {
	authConfig map[string]*string
	authURL    string
	tpl        *template.Template
	providers  map[string]provider.Provider
}

// NewApp creates new App from dependencies and Flags' config
func NewApp(authConfig map[string]*string, providers map[string]provider.Provider) *App {
	app := &App{
		authConfig: authConfig,
		authURL:    *authConfig[`url`],
		tpl:        template.Must(template.New(`iot`).ParseGlob(`./web/*.gohtml`)),
		providers:  providers,
	}

	for _, provider := range providers {
		provider.SetRenderer(app)
	}

	return app
}

// RenderDashboard render dashboard
func (a *App) RenderDashboard(w http.ResponseWriter, r *http.Request, status int, message *provider.Message) {
	response := map[string]interface{}{
		`Message`: message,
	}

	for name, provider := range a.providers {
		response[name] = provider.GetData()
	}

	if err := httputils.WriteHTMLTemplate(a.tpl.Lookup(`iot`), w, response, status); err != nil {
		httputils.InternalServerError(w, err)
	}
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return auth.HandlerWithFail(a.authConfig, func(w http.ResponseWriter, r *http.Request, _ *auth.User) {
		a.RenderDashboard(w, r, http.StatusOK, nil)
	}, func(w http.ResponseWriter, r *http.Request, err error) {
		handleAuthFail(w, r, err, a.authURL)
	})
}

func handleAuthFail(w http.ResponseWriter, r *http.Request, err error, authURL string) {
	if auth.IsForbiddenErr(err) {
		httputils.Forbidden(w)
	} else if err == auth.ErrEmptyAuthorization {
		http.Redirect(w, r, path.Join(authURL, `/redirect/github`), http.StatusFound)
	} else {
		httputils.Unauthorized(w, err)
	}
}
