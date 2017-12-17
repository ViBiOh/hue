package iot

import (
	"bytes"
	"html/template"
	"net/http"
	"path"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

type config struct {
	IFTTTSecureKey string
}

var (
	url            string
	users          map[string]*auth.User
	tpl            *template.Template
	templateConfig *config
	minifier       *minify.M
)

// Init handler
func Init(authConfig map[string]*string, iftttSecureKey string) error {
	url = *authConfig[`url`]
	users = auth.LoadUsersProfiles(*authConfig[`users`])

	tpl = template.Must(template.New(`iot`).ParseGlob(`./web/*.gohtml`))
	templateConfig = &config{
		IFTTTSecureKey: iftttSecureKey,
	}

	minifier = minify.New()
	minifier.AddFunc(`text/css`, css.Minify)
	minifier.AddFunc(`text/html`, html.Minify)

	return nil
}

func writeHTMLTemplate(w http.ResponseWriter, templateName string, content *config) error {
	templateBuffer := &bytes.Buffer{}
	if err := tpl.ExecuteTemplate(templateBuffer, templateName, content); err != nil {
		return err
	}

	w.Header().Add(`Content-Type`, `text/html; charset=UTF-8`)
	minifier.Minify(`text/html`, w, templateBuffer)
	return nil
}

// Handler for IOT request. Should be use with net/http
func Handler() http.Handler {
	return auth.HandlerWithFail(url, users, func(w http.ResponseWriter, r *http.Request, _ *auth.User) {
		if err := writeHTMLTemplate(w, `iot`, templateConfig); err != nil {
			httputils.InternalServerError(w, err)
		}
	}, func(w http.ResponseWriter, r *http.Request, err error) {
		if auth.IsForbiddenErr(err) {
			httputils.Forbidden(w)
		} else if err == auth.ErrEmptyAuthorization {
			http.Redirect(w, r, path.Join(url, `/redirect/github`), http.StatusFound)
		} else {
			httputils.Unauthorized(w, err)
		}
	})
}
