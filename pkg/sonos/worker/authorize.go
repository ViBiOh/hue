package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/sonos"
)

const (
	refreshTokenURL = "https://api.sonos.com/login/v3/oauth/access"
)

func (a *App) refreshAccessToken(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	payload := url.Values{
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{a.refreshToken},
	}

	headers := http.Header{
		"Authorization": []string{request.GenerateBasicAuth(a.clientID, a.clientSecret)},
	}

	body, _, _, err := request.PostForm(ctx, refreshTokenURL, payload, headers)
	if err != nil {
		return err
	}

	rawData, err := request.ReadBody(body)
	if err != nil {
		return err
	}

	var token sonos.Token
	if err := json.Unmarshal(rawData, &token); err != nil {
		return errors.WithStack(err)
	}

	a.accessToken = token.AccessToken

	return nil
}

func (a *App) requestWithAuth(ctx context.Context, req *http.Request) ([]byte, error) {
	if req.Header == nil {
		req.Header = http.Header{}
	}

	a.mutex.RLock()
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.accessToken))
	a.mutex.RUnlock()

	body, status, _, err := request.DoAndRead(ctx, req)
	if err != nil {
		if status == http.StatusUnauthorized {
			if err := a.refreshAccessToken(ctx); err != nil {
				return nil, err
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.accessToken))

			body, _, _, err = request.DoAndRead(ctx, req)
		}

		if err != nil {
			return nil, err
		}
	}

	return request.ReadBody(body)
}
