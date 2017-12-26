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
func RenderDashboard(w http.ResponseWriter, r *http.Request, tpl *template.Template, netatmoClient *netatmo.Client, message *message) {
	netatmoData, err := netatmoClient.GetStationData()
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

func handleAuthSuccess(w http.ResponseWriter, r *http.Request, tpl *template.Template, netatmoClient *netatmo.Client) {
	values := r.URL.Query()
	messageContent := values.Get(`message_content`)

	if messageContent != `` {
		RenderDashboard(w, r, tpl, netatmoClient, &message{Level: values.Get(`message_level`), Content: messageContent})
	} else {
		RenderDashboard(w, r, tpl, netatmoClient, nil)
	}
}

// NewHandler create Handler from Flags' config
func NewHandler(authConfig map[string]*string, netatmoClient *netatmo.Client) http.Handler {
	authURL := *authConfig[`url`]
	users := auth.LoadUsersProfiles(*authConfig[`users`])
	tpl := template.Must(template.New(`iot`).ParseGlob(`./web/*.gohtml`))

	return auth.HandlerWithFail(authURL, users, func(w http.ResponseWriter, r *http.Request, _ *auth.User) {
		handleAuthSuccess(w, r, tpl, netatmoClient)
	}, func(w http.ResponseWriter, r *http.Request, err error) {
		handleAuthFail(w, r, err, authURL)
	})
}
