package sonos

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	sonosAPIURL       = `https://api.sonos.com`
	sonosRefreshToken = `/login/v3/oauth/access`
)

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
		`accessToken`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sAccessToken`, prefix)), ``, `[sonos] Access Token`),
		`refreshToken`: flag.String(tools.ToCamel(fmt.Sprintf(`%sRefreshToken`, prefix)), ``, `[sonos] Refresh Token`),
		`clientID`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sClientID`, prefix)), ``, `[sonos] Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[sonos] Client Secret`),
	}
}

func (a *App) refreshAccessToken(ctx context.Context) error {
	payload := fmt.Sprintf(`grant_type=refresh_token&refresh_token=%s`, a.refreshToken)
	rawData, err := request.Do(
		ctx,
		fmt.Sprintf(`%s%s`, sonosAPIURL, sonosRefreshToken),
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

// SetHub receive Hub during init of it
func (a *App) SetHub(provider.Hub) {
}

// GetWorkerSource get source of message in websocket
func (a *App) GetWorkerSource() string {
	return `sonos`
}

// GetData return data for Dashboard rendering
func (a *App) GetData(ctx context.Context) interface{} {
	data, err := a.ListHouseholds(ctx)
	if err != nil {
		rollbar.LogError(`[sonos] Error while listing households: %v`, err)
	}

	return data
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	return fmt.Errorf(`Unknown worker command: %s`, message.Type)
}

// Handler for request. Should be use with net/http
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if err := httpjson.ResponseJSON(w, http.StatusOK, a.GetData(r.Context()), httpjson.IsPretty(r)); err != nil {
				httperror.InternalServerError(w, err)
			}
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
