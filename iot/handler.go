package iot

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/iot/netatmo"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

type message struct {
	Level   string
	Content string
}

var (
	url      string
	users    map[string]*auth.User
	tpl      *template.Template
	minifier *minify.M
)

// Init handler
func Init(authConfig map[string]*string) error {
	url = *authConfig[`url`]
	users = auth.LoadUsersProfiles(*authConfig[`users`])

	tpl = template.Must(template.New(`iot`).ParseGlob(`./web/*.gohtml`))

	minifier = minify.New()
	minifier.AddFunc(`text/css`, css.Minify)
	minifier.AddFunc(`text/html`, html.Minify)

	return nil
}

func writeHTMLTemplate(w http.ResponseWriter, templateName string, content interface{}) error {
	templateBuffer := &bytes.Buffer{}
	if err := tpl.ExecuteTemplate(templateBuffer, templateName, content); err != nil {
		return err
	}

	w.Header().Add(`Content-Type`, `text/html; charset=UTF-8`)
	minifier.Minify(`text/html`, w, templateBuffer)
	return nil
}

// RenderDashboard render dashboard
func RenderDashboard(w http.ResponseWriter, r *http.Request, message *message) {
	netatmoData, err := netatmo.GetStationData()
	if err != nil {
		log.Printf(`Error while reading Netatmo data: %v`, err)
	}

	response := map[string]interface{}{
		`Netatmo`: netatmoData,
		`Message`: message,
	}

	if err := writeHTMLTemplate(w, `iot`, response); err != nil {
		httputils.InternalServerError(w, err)
	}
}

func handleAuthFail(w http.ResponseWriter, r *http.Request, err error) {
	if auth.IsForbiddenErr(err) {
		httputils.Forbidden(w)
	} else if err == auth.ErrEmptyAuthorization {
		http.Redirect(w, r, path.Join(url, `/redirect/github`), http.StatusFound)
	} else {
		httputils.Unauthorized(w, err)
	}
}

func handleAuthSuccess(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	values := r.URL.Query()
	messageContent := values.Get(`message_content`)

	if messageContent != `` {
		RenderDashboard(w, r, &message{Level: values.Get(`message_level`), Content: messageContent})
	} else {
		RenderDashboard(w, r, nil)
	}
}

// Handler for IOT request. Should be use with net/http
func Handler() http.Handler {
	return auth.HandlerWithFail(url, users, handleAuthSuccess, handleAuthFail)
}
