package sonos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/iot/pkg/provider"
)

var (
	_ provider.WorkerProvider = &App{}
)

// App of package
type App struct {
	hub        provider.Hub
	households []*Household
	mutex      sync.RWMutex
}

// New creates new App
func New() *App {
	return &App{
		households: nil,
	}
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
		httperror.BadRequest(w, fmt.Errorf("volume is not an integer: %v", err))
		return
	}

	payload := fmt.Sprintf("%s|%d", urlParts[0], volume)

	output := a.hub.SendToWorker(r.Context(), nil, Source, VolumeAction, payload, true)
	if output.Action == provider.WorkerErrorAction {
		httperror.InternalServerError(w, fmt.Errorf("%s", output.Payload))
		return
	}
}

func (a *App) muteHandler(w http.ResponseWriter, r *http.Request, urlParts []string) {
	payload := fmt.Sprintf("%s|%t", urlParts[0], urlParts[1] == "mute")

	output := a.hub.SendToWorker(r.Context(), nil, Source, MuteAction, payload, true)
	if output.Action == provider.WorkerErrorAction {
		a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: "error", Content: fmt.Sprintf("%s", output.Payload)})
		return
	}

	a.hub.RenderDashboard(w, r, http.StatusOK, &provider.Message{Level: "success", Content: "Mute state changed"})
}

func (a *App) groupsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			body, err := request.ReadBodyRequest(r)
			if err != nil {
				httperror.InternalServerError(w, err)
				return
			}

			urlParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

			if len(urlParts) == 2 {
				if urlParts[1] == VolumeAction {
					a.volumeHandler(w, r, urlParts, body)
					return
				}

				if urlParts[1] == MuteAction {
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
	strippedGroupsHandler := http.StripPrefix("/groups", a.groupsHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/groups") {
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
	if p.Action == "households" {
		return a.handleHouseholdsWorker(p)
	}

	if p.Action == "mute" {
		return a.handleMuteWorker(p)
	}

	return provider.ErrWorkerUnknownAction
}

func (a *App) handleHouseholdsWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var data []*Household
	if err := json.Unmarshal([]byte(message.Payload), &data); err != nil {
		return err
	}

	a.households = data

	return nil
}

func (a *App) handleMuteWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	parts := strings.Split(message.Payload, "|")
	if len(parts) == 2 {
		muted, err := strconv.ParseBool(parts[1])
		if err != nil {
			return err
		}

		for _, household := range a.households {
			for _, group := range household.Groups {
				if group.ID == parts[0] {
					group.Volume.Muted = muted

					return nil
				}
			}
		}
	}

	return nil
}
