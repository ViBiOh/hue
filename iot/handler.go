package iot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

type netatmo struct {
	Body struct {
		Devices []struct {
			StationName   string `json:"station_name"`
			DashboardData struct {
				Temperature float32
				Humidity    float32
			} `json:"dashboard_data"`
			Modules []struct {
				ModuleName    string `json:"module_name"`
				DashboardData struct {
					Temperature float32
					Humidity    float32
				} `json:"dashboard_data"`
			} `json:"modules"`
		} `json:"devices"`
	} `json:"body"`
}

type config struct {
	IFTTTSecureKey string
	NetatmoToken   string
}

type response struct {
	Config  *config
	Netatmo *netatmo
}

var (
	url            string
	users          map[string]*auth.User
	tpl            *template.Template
	templateConfig *config
	minifier       *minify.M
)

// Init handler
func Init(authConfig map[string]*string, iftttSecureKey string, netatmoToken string) error {
	url = *authConfig[`url`]
	users = auth.LoadUsersProfiles(*authConfig[`users`])

	tpl = template.Must(template.New(`iot`).ParseGlob(`./web/*.gohtml`))
	templateConfig = &config{
		IFTTTSecureKey: iftttSecureKey,
		NetatmoToken:   netatmoToken,
	}

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

func getNetatmoInfo() (*netatmo, error) {
	var infos netatmo

	if rawData, err := httputils.GetBody(`https://api.netatmo.com/api/getstationsdata?access_token=`+templateConfig.NetatmoToken, nil); err != nil {
		return nil, fmt.Errorf(`Error while reading station data: %v`, err)
	} else {
		if err := json.Unmarshal(rawData, &infos); err != nil {
			return nil, fmt.Errorf(`Error while unmarshalling data: %v`, err)
		}
	}

	return &infos, nil
}

// Handler for IOT request. Should be use with net/http
func Handler() http.Handler {
	return auth.HandlerWithFail(url, users, func(w http.ResponseWriter, r *http.Request, _ *auth.User) {
		netatmoData, err := getNetatmoInfo()
		if err != nil {
			log.Printf(`Error while reading Netatmo data: %v`, err)
		}

		if err := writeHTMLTemplate(w, `iot`, &response{Config: templateConfig, Netatmo: netatmoData}); err != nil {
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
