package sonos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/provider"
)

// App stores informations and secret of API
type App struct {
	hub        provider.Hub
	households []*Household
	mutex      sync.RWMutex
}

// NewApp create Client from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		households: nil,
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return nil
}

// SetHub receive Hub during init of it
func (a *App) SetHub(hub provider.Hub) {
	a.hub = hub
}

// GetData return data for Dashboard rendering
func (a *App) GetData() interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.households
}

func (a *App) volumeHandler(w http.ResponseWriter, r *http.Request, urlParts []string, body []byte) {
	volume, err := strconv.Atoi(string(body))
	if err != nil {
		httperror.BadRequest(w, fmt.Errorf(`volume is not an integer: %v`, err))
		return
	}

	payload := fmt.Sprintf(`%s|%d`, urlParts[0], volume)

	output := a.hub.SendToWorker(r.Context(), nil, Source, VolumeAction, payload, true)
	if output.Action == provider.WorkerErrorAction {
		httperror.InternalServerError(w, fmt.Errorf(`error while setting volume of group %s: %v`, urlParts[0], output.Payload))
		return
	}
}

func (a *App) muteHandler(w http.ResponseWriter, r *http.Request, urlParts []string) {
	payload := fmt.Sprintf(`%s|%t`, urlParts[0], urlParts[1] == `mute`)

	output := a.hub.SendToWorker(r.Context(), nil, Source, MuteAction, payload, true)
	if output.Action == provider.WorkerErrorAction {
		httperror.InternalServerError(w, fmt.Errorf(`error while changing mute of group %s: %v`, urlParts[0], output.Payload))
		return
	}
}

func (a *App) groupsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			body, err := request.ReadBodyRequest(r)
			if err != nil {
				httperror.InternalServerError(w, fmt.Errorf(`error while reading body: %v`, err))
				return
			}

			urlParts := strings.Split(strings.Trim(r.URL.Path, `/`), `/`)

			if len(urlParts) == 2 {
				if urlParts[1] == VolumeAction {
					a.volumeHandler(w, r, urlParts, body)
					return
				}

				if strings.Contains(urlParts[1], MuteAction) {
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

// GetWorkerSource returns source for worker
func (a *App) GetWorkerSource() string {
	return Source
}

// WorkerHandler handler worker requests
func (a *App) WorkerHandler(p *provider.WorkerMessage) error {
	if p.Action == `households` {
		return a.handleHouseholdsWorker(p)
	}

	return provider.ErrWorkerUnknownAction
}

func (a *App) handleHouseholdsWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var data []*Household

	convert, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf(`error while converting households payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &data); err != nil {
		return fmt.Errorf(`error while unmarshalling households: %v`, err)
	}

	a.households = data

	return nil
}
