package netatmo

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
)

// StationData contains data retrieved when getting stations datas
type StationData struct {
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
					Noise       float32
					CO2         float32
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

var (
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
)

const netatmoGetStationDataURL = `https://api.netatmo.com/api/getstationsdata?access_token=`
const netatmoRefreshTokenURL = `https://api.netatmo.com/oauth2/token`

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`accessToken`:  flag.String(tools.ToCamel(prefix+`AccessToken`), ``, `[netatmo] Access Token`),
		`refreshToken`: flag.String(tools.ToCamel(prefix+`RefreshToken`), ``, `[netatmo] Refresh Token`),
		`clientID`:     flag.String(tools.ToCamel(prefix+`ClientID`), ``, `[netatmo] Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(prefix+`ClientSecret`), ``, `[netatmo] Client Secret`),
	}
}

// Init retrieves config from CLI args
func Init(config map[string]*string) error {
	clientID = *config[`clientID`]
	clientSecret = *config[`clientSecret`]
	accessToken = *config[`accessToken`]
	refreshToken = *config[`refreshToken`]

	return nil
}

func refreshAccessToken() error {
	log.Print(`Refreshing Netatmo Access Token`)

	rawData, err := httputils.PostBody(netatmoRefreshTokenURL, []byte(`grant_type=refresh_token&refresh_token=`+refreshToken+`&client_id=`+clientID+`&client_secret=`+clientSecret), map[string]string{`Content-Type`: `application/x-www-form-urlencoded;charset=UTF-8`})

	if err != nil {
		return fmt.Errorf(`Error while refreshing token: %v`, err)
	}

	var token netatmoToken
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`Error while unmarshalling token: %v`, err)
	}

	accessToken = token.AccessToken

	return nil
}

// GetStationData retrieves Station data of user
func GetStationData() (*StationData, error) {
	if accessToken == `` {
		return nil, nil
	}

	var infos StationData

	rawData, err := httputils.GetBody(netatmoGetStationDataURL+accessToken, nil)
	if err != nil {
		var netatmoErrorValue netatmoError

		if err := json.Unmarshal(rawData, &netatmoErrorValue); err != nil {
			return nil, fmt.Errorf(`Error while unmarshalling error: %v`, err)
		}

		if netatmoErrorValue.Error.Code == 3 {
			if err := refreshAccessToken(); err != nil {
				return nil, fmt.Errorf(`Error while refreshing access token: %v`, err)
			}
			return GetStationData()
		}

		return nil, fmt.Errorf(`Error while getting data: %v`, err)
	}

	if err := json.Unmarshal(rawData, &infos); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling data: %v`, err)
	}

	return &infos, nil
}
