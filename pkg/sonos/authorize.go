package sonos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/pkg/request"
)

const (
	refreshTokenURL = `https://api.sonos.com/login/v3/oauth/access`
)

func (a *App) refreshAccessToken(ctx context.Context) error {
	payload := url.Values{
		`grant_type`:    []string{`refresh_token`},
		`refresh_token`: []string{a.refreshToken},
	}

	headers := http.Header{
		`Authorization`: []string{request.GenerateBasicAuth(a.clientID, a.clientSecret)},
	}

	rawData, err := request.PostForm(ctx, refreshTokenURL, payload, headers)
	if err != nil {
		return fmt.Errorf(`error while refreshing token: %v`, err)
	}

	var token refreshToken
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`error while unmarshalling token %s: %v`, rawData, err)
	}

	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()
	a.accessToken = token.AccessToken

	return nil
}

func (a *App) requestWithAuth(ctx context.Context, req *http.Request) ([]byte, error) {
	if req.Header == nil {
		req.Header = http.Header{}
	}
	req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, a.accessToken))

	data, err := request.DoAndRead(ctx, req)

	if err != nil && strings.Contains(string(data), `Incorrect token`) {
		if err := a.refreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf(`error while refreshing access token: %v`, err)
		}

		req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, a.accessToken))
		return request.DoAndRead(ctx, req)
	}

	return data, err
}
