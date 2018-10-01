package netatmo

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	netatmoGetStationDataURL = `https://api.netatmo.com/api/getstationsdata?access_token=`
	netatmoRefreshTokenURL   = `https://api.netatmo.com/oauth2/token`
)

// App stores informations and secret of API
type App struct {
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
}

// NewApp create Client from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		clientID:     *config[`clientID`],
		clientSecret: *config[`clientSecret`],
		accessToken:  *config[`accessToken`],
		refreshToken: *config[`refreshToken`],
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`accessToken`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sAccessToken`, prefix)), ``, `[netatmo] Access Token`),
		`refreshToken`: flag.String(tools.ToCamel(fmt.Sprintf(`%sRefreshToken`, prefix)), ``, `[netatmo] Refresh Token`),
		`clientID`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sClientID`, prefix)), ``, `[netatmo] Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[netatmo] Client Secret`),
	}
}

func (a *App) refreshAccessToken(ctx context.Context) error {
	payload := url.Values{
		`grant_type`:    []string{`refresh_token`},
		`refresh_token`: []string{a.refreshToken},
		`client_id`:     []string{a.clientID},
		`client_secret`: []string{a.clientSecret},
	}

	rawData, err := request.PostForm(ctx, netatmoRefreshTokenURL, payload, nil)
	if err != nil {
		return fmt.Errorf(`error while refreshing token: %v`, err)
	}

	var token netatmoToken
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`error while unmarshalling token %s: %v`, rawData, err)
	}

	a.accessToken = token.AccessToken

	return nil
}

func (a *App) getStationData(ctx context.Context) (*StationData, error) {
	if a.accessToken == `` {
		return nil, nil
	}

	rawData, err := request.Get(ctx, fmt.Sprintf(`%s%s`, netatmoGetStationDataURL, a.accessToken), nil)
	if err != nil {
		var netatmoErrorValue netatmoError

		if err := json.Unmarshal(rawData, &netatmoErrorValue); err != nil {
			return nil, fmt.Errorf(`error while unmarshalling error %s: %v`, rawData, err)
		}

		if netatmoErrorValue.Error.Code == 3 || netatmoErrorValue.Error.Code == 2 {
			if err := a.refreshAccessToken(ctx); err != nil {
				return nil, fmt.Errorf(`error while refreshing access token: %v`, err)
			}
			return a.getStationData(ctx)
		}

		return nil, fmt.Errorf(`error while getting data: %v`, err)
	}

	var infos StationData
	if err := json.Unmarshal(rawData, &infos); err != nil {
		return nil, fmt.Errorf(`error while unmarshalling data %s: %v`, rawData, err)
	}

	return &infos, nil
}

// SetHub receive Hub during init of it
func (a *App) SetHub(provider.Hub) {
}

// GetWorkerSource get source of message in websocket
func (a *App) GetWorkerSource() string {
	return `netatmo`
}

// GetData return data for Dashboard rendering
func (a *App) GetData(ctx context.Context) interface{} {
	return true
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	return fmt.Errorf(`unknown worker command: %s`, message.Type)
}

// Handler for request. Should be use with net/http
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			data, err := a.getStationData(r.Context())
			if err != nil {
				httperror.InternalServerError(w, fmt.Errorf(`error while getting station data: %v`, err))
				return
			}

			if err := httpjson.ResponseJSON(w, http.StatusOK, data, httpjson.IsPretty(r)); err != nil {
				httperror.InternalServerError(w, err)
				return
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}
