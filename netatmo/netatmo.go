package netatmo

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/provider"
)

const (
	netatmoGetStationDataURL = `https://api.netatmo.com/api/getstationsdata?access_token=`
	netatmoRefreshTokenURL   = `https://api.netatmo.com/oauth2/token`
)

var (
	// WebSocketPrefix ws message prefix for all hue commands
	WebSocketPrefix = []byte(`netatmo `)
)

// StationData contains data retrieved when getting stations datas
type StationData struct {
	Body struct {
		Devices []struct {
			StationName   string `json:"station_name"`
			DashboardData struct {
				Temperature float32
				Humidity    float32
				Noise       float32
				CO2         float32
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

type netatmoError struct {
	Error struct {
		Code    int
		Message string
	}
}

type netatmoToken struct {
	AccessToken string `json:"access_token"`
}

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
		`accessToken`:  flag.String(tools.ToCamel(prefix+`AccessToken`), ``, `[netatmo] Access Token`),
		`refreshToken`: flag.String(tools.ToCamel(prefix+`RefreshToken`), ``, `[netatmo] Refresh Token`),
		`clientID`:     flag.String(tools.ToCamel(prefix+`ClientID`), ``, `[netatmo] Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(prefix+`ClientSecret`), ``, `[netatmo] Client Secret`),
	}
}

func (a *App) refreshAccessToken() error {
	rawData, err := httputils.Request(netatmoRefreshTokenURL, []byte(fmt.Sprintf(`grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s`, a.refreshToken, a.clientID, a.clientSecret)), map[string]string{`Content-Type`: `application/x-www-form-urlencoded;charset=UTF-8`}, http.MethodPost)

	if err != nil {
		return fmt.Errorf(`[netatmo] Error while refreshing token: %v`, err)
	}

	var token netatmoToken
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`[netatmo] Error while unmarshalling token: %v`, err)
	}

	a.accessToken = token.AccessToken

	return nil
}

// GetStationData retrieves Station data of user
func (a *App) GetStationData() (*StationData, error) {
	if a.accessToken == `` {
		return nil, nil
	}

	var infos StationData

	rawData, err := httputils.GetRequest(netatmoGetStationDataURL+a.accessToken, nil)
	if err != nil {
		var netatmoErrorValue netatmoError

		if err := json.Unmarshal(rawData, &netatmoErrorValue); err != nil {
			return nil, fmt.Errorf(`[netatmo] Error while unmarshalling error: %v`, err)
		}

		if netatmoErrorValue.Error.Code == 3 || netatmoErrorValue.Error.Code == 2 {
			if err := a.refreshAccessToken(); err != nil {
				return nil, fmt.Errorf(`[netatmo] Error while refreshing access token: %v`, err)
			}
			return a.GetStationData()
		}

		return nil, fmt.Errorf(`[netatmo] Error while getting data: %v`, err)
	}

	if err := json.Unmarshal(rawData, &infos); err != nil {
		return nil, fmt.Errorf(`[netatmo] Error while unmarshalling data: %v`, err)
	}

	return &infos, nil
}

// SetHub receive Hub during init of it
func (a *App) SetHub(provider.Hub) {
}

// GetWorkerPrefix get prefix of message in websocket
func (a *App) GetWorkerPrefix() []byte {
	return WebSocketPrefix
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
func (a *App) WorkerHandler(payload []byte) {
	log.Printf(`[netatmo] Unknown worker command: %s`, payload)
}
