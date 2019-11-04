package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/v3/pkg/request"
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

	resp, err := request.New().Post(refreshTokenURL).BasicAuth(a.clientID, a.clientSecret).Form(ctx, payload)
	if err != nil {
		return err
	}

	rawData, err := request.ReadBodyResponse(resp)
	if err != nil {
		return err
	}

	var token sonos.Token
	if err := json.Unmarshal(rawData, &token); err != nil {
		return err
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

	resp, err := request.Do(req)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			if err := a.refreshAccessToken(ctx); err != nil {
				return nil, err
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.accessToken))

			resp, err = request.Do(req)
		}

		if err != nil {
			return nil, err
		}
	}

	return request.ReadBodyResponse(resp)
}
