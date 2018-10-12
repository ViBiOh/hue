package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/sonos"
)

const (
	refreshTokenURL = `https://api.sonos.com/login/v3/oauth/access`
)

func (a *App) refreshAccessToken(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

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

	var token sonos.Token
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`error while unmarshalling token %s: %v`, rawData, err)
	}

	a.accessToken = token.AccessToken

	return nil
}

func (a *App) requestWithAuth(ctx context.Context, req *http.Request) ([]byte, error) {
	if req.Header == nil {
		req.Header = http.Header{}
	}

	a.mutex.RLock()
	req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, a.accessToken))
	a.mutex.RUnlock()

	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf(`error while reading body: %v`, err)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
	data, err := request.DoAndRead(ctx, req)

	if err != nil && strings.Contains(string(data), `Incorrect token`) {
		if err := a.refreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf(`error while refreshing access token: %v`, err)
		}

		req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, a.accessToken))

		req.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		return request.DoAndRead(ctx, req)
	}

	return data, err
}
