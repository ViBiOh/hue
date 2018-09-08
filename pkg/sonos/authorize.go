package sonos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/request"
)

const (
	refreshTokenURL = `https://api.sonos.com/login/v3/oauth/access`
)

func (a *App) refreshAccessToken(ctx context.Context) error {
	payload := fmt.Sprintf(`grant_type=refresh_token&refresh_token=%s`, a.refreshToken)
	rawData, err := request.Do(
		ctx,
		refreshTokenURL,
		[]byte(payload),
		http.Header{
			`Authorization`: []string{request.GetBasicAuth(a.clientID, a.clientSecret)},
			`Content-Type`:  []string{`application/x-www-form-urlencoded;charset=UTF-8`},
		},
		http.MethodPost,
	)

	if err != nil {
		return fmt.Errorf(`Error while refreshing token: %v`, err)
	}

	var token refreshToken
	if err := json.Unmarshal(rawData, &token); err != nil {
		return fmt.Errorf(`Error while unmarshalling token %s: %v`, rawData, err)
	}

	a.accessToken = token.AccessToken

	return nil
}

func (a *App) requestWithAuth(ctx context.Context, req *http.Request) ([]byte, error) {
	req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, a.accessToken))

	return request.DoAndRead(ctx, req)
}
