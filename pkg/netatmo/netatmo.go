package netatmo

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

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
		`accessToken`:  flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `AccessToken`)), ``, `[netatmo] Access Token`),
		`refreshToken`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `RefreshToken`)), ``, `[netatmo] Refresh Token`),
		`clientID`:     flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `ClientID`)), ``, `[netatmo] Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `ClientSecret`)), ``, `[netatmo] Client Secret`),
	}
}

func (a *App) refreshAccessToken() error {
	payload := fmt.Sprintf(`grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s`, a.refreshToken, a.clientID, a.clientSecret)
	rawData, err := request.Do(netatmoRefreshTokenURL, []byte(payload), map[string]string{`Content-Type`: `application/x-www-form-urlencoded;charset=UTF-8`}, http.MethodPost)

	if err != nil {
		return fmt.Errorf(`Error while refreshing token: %v`, err)
	}

	var token netatmoToken
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`Error while unmarshalling token %s: %v`, rawData, err)
	}

	a.accessToken = token.AccessToken

	return nil
}

// GetStationData retrieves Station data of user
func (a *App) GetStationData() (*StationData, error) {
	if a.accessToken == `` {
		return nil, nil
	}

	rawData, err := request.Get(fmt.Sprintf(`%s%s`, netatmoGetStationDataURL, a.accessToken), nil)
	if err != nil {
		var netatmoErrorValue netatmoError

		if err := json.Unmarshal(rawData, &netatmoErrorValue); err != nil {
			return nil, fmt.Errorf(`Error while unmarshalling error %s: %v`, rawData, err)
		}

		if netatmoErrorValue.Error.Code == 3 || netatmoErrorValue.Error.Code == 2 {
			if err := a.refreshAccessToken(); err != nil {
				return nil, fmt.Errorf(`Error while refreshing access token: %v`, err)
			}
			return a.GetStationData()
		}

		return nil, fmt.Errorf(`Error while getting data: %v`, err)
	}

	var infos StationData
	if err := json.Unmarshal(rawData, &infos); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data %s: %v`, rawData, err)
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
func (a *App) GetData() interface{} {
	data, err := a.GetStationData()
	if err != nil {
		log.Printf(`[netatmo] Error while getting station data: %v`, err)
	}

	return data
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	return fmt.Errorf(`Unknown worker command: %s`, message.Type)
}
