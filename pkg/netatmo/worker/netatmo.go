package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ViBiOh/httputils/pkg/errors"
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

	body, _, _, err := request.PostForm(ctx, netatmoRefreshTokenURL, payload, nil)
	if err != nil {
		return err
	}

	rawData, err := request.ReadBody(body)
	if err != nil {
		return err
	}

	var token netatmo.Token
	if err := json.Unmarshal(rawData, &token); err != nil {
		return errors.WithStack(err)
	}

	a.accessToken = token.AccessToken

	return nil
}

func isInvalidTokenError(rawData []byte, err error) bool {
	var netatmoErrorValue netatmo.Error

	if err := json.Unmarshal(rawData, &netatmoErrorValue); err != nil {
		logger.Error(`%+v`, errors.WithStack(err))
		return false
	}

	return netatmoErrorValue.Error.Code == 3 || netatmoErrorValue.Error.Code == 2
}

func (a *App) getStationsData(ctx context.Context, retry bool) (*netatmo.StationsData, error) {
	if a.accessToken == `` {
		return nil, nil
	}

	a.mutex.RLock()
	body, _, _, err := request.Get(ctx, fmt.Sprintf(`%s%s`, netatmoGetStationsDataURL, a.accessToken), nil)
	a.mutex.RUnlock()

	if err != nil {
		return nil, err
	}

	rawData, err := request.ReadBody(body)
	if err != nil {
		return nil, err
	}

	if err != nil {
		if isInvalidTokenError(rawData, err) && retry {
			if err := a.refreshAccessToken(ctx); err != nil {
				return nil, err
			}

			return a.getStationsData(ctx, false)
		}

		return nil, err
	}

	var infos netatmo.StationsData
	if err := json.Unmarshal(rawData, &infos); err != nil {
		return nil, errors.WithStack(err)
	}

	return &infos, nil
}
