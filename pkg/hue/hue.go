package hue

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	groupsRequest    = `/groups`
	schedulesRequest = `/schedules`
)

// App stores informations and secret of API
type App struct {
	hub       provider.Hub
	groups    map[string]*Group
	scenes    map[string]*Scene
	schedules map[string]*Schedule
	mutex     sync.RWMutex
}

// New creates new App
func New() *App {
	return &App{}
}

func (a *App) sendWorkerMessage(w http.ResponseWriter, r *http.Request, payload string, typeName, successMessage string) {
	output := a.hub.SendToWorker(r.Context(), nil, Source, typeName, payload, true)

	if output == nil {
		a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{
			Level:   `error`,
			Content: fmt.Sprintf(`[%s] Timeout while sending message %s to Worker`, Source, typeName),
		})
	} else if output.Action == provider.WorkerErrorAction {
		a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{
			Level:   `error`,
			Content: fmt.Sprintf(`[%s] Error while sending message %s to worker: %v`, Source, typeName, output.Payload),
		})
	} else {
		a.hub.RenderDashboard(w, r, http.StatusOK, &provider.Message{
			Level:   `success`,
			Content: successMessage,
		})
	}
}

func (a *App) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		postMethod := r.FormValue(`method`)

		if postMethod == http.MethodPost {
			config := &ScheduleConfig{
				Name:      r.FormValue(`name`),
				Group:     r.FormValue(`group`),
				Localtime: ComputeScheduleReccurence(r.Form[`days[]`], r.FormValue(`hours`), r.FormValue(`minutes`)),
				State:     r.FormValue(`state`),
			}

			payload, err := json.Marshal(config)
			if err != nil {
				a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[%s] Error while marshalling schedule config: %v`, Source, err)})
				return
			}

			a.sendWorkerMessage(w, r, fmt.Sprintf(`%s`, payload), fmt.Sprintf(`%s/%s`, WorkerSchedulesAction, CreateAction), fmt.Sprintf(`%s schedule has been created`, config.Name))
			return
		}

		id := strings.Trim(strings.TrimPrefix(r.URL.Path, schedulesRequest), `/`)

		if postMethod == http.MethodPatch {
			schedule := &Schedule{
				ID: id,
				APISchedule: &APISchedule{
					Status: r.FormValue(`status`),
				},
			}

			payload, err := json.Marshal(schedule)
			if err != nil {
				a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[%s] Error while marshalling schedule: %v`, Source, err)})
				return
			}

			a.sendWorkerMessage(w, r, fmt.Sprintf(`%s`, payload), fmt.Sprintf(`%s/%s`, WorkerSchedulesAction, UpdateAction), fmt.Sprintf(`%s schedule has been %s`, r.FormValue(`name`), schedule.Status))
			return
		}

		if postMethod == http.MethodDelete {
			a.sendWorkerMessage(w, r, id, fmt.Sprintf(`%s/%s`, WorkerSchedulesAction, DeleteAction), fmt.Sprintf(`%s schedule has been deleted`, r.FormValue(`name`)))
			return
		}
	}

	a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[%s] Unknown schedule command`, Source)})
}

func (a *App) handleGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		postMethod := r.FormValue(`method`)

		if postMethod == http.MethodPatch {
			group := strings.Trim(strings.TrimPrefix(r.URL.Path, groupsRequest), `/`)
			state := r.FormValue(`state`)

			groupObj, ok := a.groups[group]
			if !ok {
				a.hub.RenderDashboard(w, r, http.StatusNotFound, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[%s] Unknown group`, Source)})
			}

			a.sendWorkerMessage(w, r, fmt.Sprintf(`%s|%s`, group, state), fmt.Sprintf(`%s/%s`, WorkerStateAction, UpdateAction), fmt.Sprintf(`%s is now %s`, groupObj.Name, state))
			return
		}
	}

	a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[%s] Unknown group command`, Source)})
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, groupsRequest) {
			a.handleGroup(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, schedulesRequest) {
			a.handleSchedule(w, r)
			return
		}

		a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[%s] Unknown command`, Source)})
	})
}

// SetHub receive Hub during init of it
func (a *App) SetHub(hub provider.Hub) {
	a.hub = hub
}

// GetWorkerSource get source of message in websocket
func (a *App) GetWorkerSource() string {
	return Source
}

// GetData return data for Dashboard rendering
func (a *App) GetData() interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if len(a.groups) == 0 && len(a.scenes) == 0 && len(a.schedules) == 0 {
		return false
	}

	return &Data{
		Groups:    a.groups,
		Scenes:    a.scenes,
		Schedules: a.schedules,
		States:    States,
	}
}
