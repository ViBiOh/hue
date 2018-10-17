package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/netatmo"
)

const (
	netatmoGetStationsDataURL = `https://api.netatmo.com/api/getstationsdata?access_token=`
	netatmoRefreshTokenURL    = `https://api.netatmo.com/oauth2/token`
)

func (a *App) refreshAccessToken(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

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

	var token netatmo.Token
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`error while unmarshalling token %s: %v`, rawData, err)
	}

	a.accessToken = token.AccessToken

	return nil
}

func isInvalidTokenError(rawData []byte, err error) bool {
	var netatmoErrorValue netatmo.Error

	if err := json.Unmarshal(rawData, &netatmoErrorValue); err != nil {
		logger.Error(`error while unmarshalling error %s: %v`, rawData, err)
		return false
	}

	return netatmoErrorValue.Error.Code == 3 || netatmoErrorValue.Error.Code == 2
}

func (a *App) getStationsData(ctx context.Context, retry bool) (*netatmo.StationsData, error) {
	if a.accessToken == `` {
		return nil, nil
	}

	a.mutex.RLock()
	rawData, err := request.Get(ctx, fmt.Sprintf(`%s%s`, netatmoGetStationsDataURL, a.accessToken), nil)
	a.mutex.RUnlock()

	if err != nil {
		if isInvalidTokenError(rawData, err) && retry {
			if err := a.refreshAccessToken(ctx); err != nil {
				return nil, fmt.Errorf(`error while refreshing access token: %v`, err)
			}

			return a.getStationsData(ctx, false)
		}

		return nil, fmt.Errorf(`error while getting data: %v`, err)
	}

	var infos netatmo.StationsData
	if err := json.Unmarshal(rawData, &infos); err != nil {
		return nil, fmt.Errorf(`error while unmarshalling data %s: %v`, rawData, err)
	}

	return &infos, nil
}
