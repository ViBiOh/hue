package sonos

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
)

// App stores informations and secret of API
type App struct {
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
	households   []*Household
	tokenMutex   sync.Mutex
}

// NewApp create Client from Flags' config
func NewApp(config map[string]*string) *App {
	app := &App{
		clientID:     *config[`clientID`],
		clientSecret: *config[`clientSecret`],
		accessToken:  *config[`accessToken`],
		refreshToken: *config[`refreshToken`],
		households:   make([]*Household, 0),
	}

	households, err := app.GetHouseholds(context.Background())
	if err != nil {
		rollbar.LogError(`[sonos] Error while listing households: %v`, err)
	} else {
		app.households = households
	}

	return app
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

// SetHub receive Hub during init of it
func (a *App) SetHub(provider.Hub) {
}

// GetWorkerSource get source of message in websocket
func (a *App) GetWorkerSource() string {
	return `sonos`
}

// GetData return data for Dashboard rendering
func (a *App) GetData(ctx context.Context) interface{} {
	groups := make([]*Group, 0)

	for _, household := range a.households {
		data, err := a.GetGroups(ctx, household.ID)
		if err != nil {
			rollbar.LogError(`[sonos] Error while listing groups: %v`, err)
		} else {
			groups = append(groups, data.Groups...)
		}
	}

	for _, group := range groups {
		data, err := a.GetGroupVolume(ctx, group.ID)
		if err != nil {
			rollbar.LogError(`[sonos] Error while getting group volume: %v`, err)
		} else {
			group.Volume = data
		}
	}

	return groups
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	return fmt.Errorf(`Unknown worker command: %s`, message.Type)
}

// Handler for request. Should be use with net/http
func (a *App) Handler() http.Handler {
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
