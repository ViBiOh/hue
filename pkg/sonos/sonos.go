package sonos

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/request"
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
	return true
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(message *provider.WorkerMessage) error {
	return fmt.Errorf(`unknown worker command: %s`, message.Type)
}

func (a *App) getGroupsData(ctx context.Context) ([]*Group, error) {
	groups := make([]*Group, 0)

	for _, household := range a.households {
		data, err := a.GetGroups(ctx, household.ID)
		if err != nil {
			return nil, fmt.Errorf(`[sonos] Error while listing groups: %v`, err)
		}

		groups = append(groups, data.Groups...)
	}

	for _, group := range groups {
		data, err := a.GetGroupVolume(ctx, group.ID)
		if err != nil {
			return nil, fmt.Errorf(`[sonos] Error while getting group volume: %v`, err)
		}

		group.Volume = data
	}

	return groups, nil
}

func (a *App) volumeHandler(w http.ResponseWriter, r *http.Request, urlParts []string, body []byte) {
	volume, err := strconv.Atoi(string(body))
	if err != nil {
		httperror.BadRequest(w, fmt.Errorf(`volume is not an integer: %v`, err))
		return
	}

	if _, err = a.SetGroupVolume(r.Context(), urlParts[0], volume); err != nil {
		httperror.InternalServerError(w, fmt.Errorf(`error while setting volume of group %s: %v`, urlParts[0], err))
		return
	}
}

func (a *App) muteHandler(w http.ResponseWriter, r *http.Request, urlParts []string) {
	if err := a.SetGroupMute(r.Context(), urlParts[0], urlParts[1] == `mute`); err != nil {
		httperror.InternalServerError(w, fmt.Errorf(`error while changing mute of group %s: %v`, urlParts[0], err))
		return
	}
}

func (a *App) groupsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			groups, err := a.getGroupsData(r.Context())
			if err != nil {
				httperror.InternalServerError(w, fmt.Errorf(`error while getting groups data: %v`, err))
				return
			}

			if err = httpjson.ResponseJSON(w, http.StatusOK, groups, httpjson.IsPretty(r)); err != nil {
				httperror.InternalServerError(w, fmt.Errorf(`error while marshalling JSON response: %v`, err))
				return
			}

		case http.MethodPost:
			body, err := request.ReadBodyRequest(r)
			if err != nil {
				httperror.InternalServerError(w, fmt.Errorf(`error while reading body: %v`, err))
				return
			}

			urlParts := strings.Split(strings.Trim(r.URL.Path, `/`), `/`)

			if len(urlParts) == 2 {
				if urlParts[1] == `volume` {
					a.volumeHandler(w, r, urlParts, body)
					return
				}

				if strings.Contains(urlParts[1], `mute`) {
					a.muteHandler(w, r, urlParts)
					return
				}
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// Handler for request. Should be use with net/http
func (a *App) Handler() http.Handler {
	strippedGroupsHandler := http.StripPrefix(`/groups`, a.groupsHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, `/groups`) {
			strippedGroupsHandler.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
