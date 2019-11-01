package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

	req, err := request.New(ctx, http.MethodPost, refreshTokenURL, strings.NewReader(payload.Encode()), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(a.clientID, a.clientSecret)

	response, err := request.Do(ctx, req)
	if err != nil {
		return err
	}

	rawData, err := request.ReadBodyResponse(response)
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

	response, err := request.Do(ctx, req)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusUnauthorized {
			if err := a.refreshAccessToken(ctx); err != nil {
				return nil, err
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.accessToken))

			response, err = request.Do(ctx, req)
		}

		if err != nil {
			return nil, err
		}
	}

	return request.ReadBodyResponse(response)
}
