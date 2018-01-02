package iot

import (
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/iot/netatmo"
)

type message struct {
	Level   string
	Content string
}

// RenderDashboard render dashboard
func RenderDashboard(w http.ResponseWriter, r *http.Request, tpl *template.Template, netatmoApp *netatmo.App, message *message) {
	netatmoData, err := netatmoApp.GetStationData()
	if err != nil {
		log.Printf(`Error while reading Netatmo data: %v`, err)
	}

	response := map[string]interface{}{
		`Netatmo`: netatmoData,
		`Message`: message,
	}

	if err := httputils.WriteHTMLTemplate(tpl.Lookup(`iot`), w, response); err != nil {
		httputils.InternalServerError(w, err)
	}
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

func handleAuthSuccess(w http.ResponseWriter, r *http.Request, tpl *template.Template, netatmoApp *netatmo.App) {
	values := r.URL.Query()
	messageContent := values.Get(`message_content`)

	if messageContent != `` {
		RenderDashboard(w, r, tpl, netatmoApp, &message{Level: values.Get(`message_level`), Content: messageContent})
	} else {
		RenderDashboard(w, r, tpl, netatmoApp, nil)
	}
}

// Handler create Handler from Flags' config
func Handler(authConfig map[string]*string, netatmoApp *netatmo.App) http.Handler {
	authURL := *authConfig[`url`]
	tpl := template.Must(template.New(`iot`).ParseGlob(`./web/*.gohtml`))

	return auth.HandlerWithFail(authConfig, func(w http.ResponseWriter, r *http.Request, _ *auth.User) {
		handleAuthSuccess(w, r, tpl, netatmoApp)
	}, func(w http.ResponseWriter, r *http.Request, err error) {
		handleAuthFail(w, r, err, authURL)
	})
}
