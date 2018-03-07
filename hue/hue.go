package hue

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/iot/provider"
	"github.com/ViBiOh/iot/utils"
)

const (
	// HueSource constant for worker message
	HueSource = `hue`

	groupsRequest    = `/groups`
	schedulesRequest = `/schedules`
)

var (
	// States available states of lights
	States = map[string]map[string]interface{}{
		`off`: {
			`on`:             false,
			`transitiontime`: 30,
		},
		`on`: {
			`on`:             true,
			`transitiontime`: 30,
			`sat`:            0,
			`bri`:            254,
		},
		`dimmed`: {
			`on`:             true,
			`transitiontime`: 30,
			`sat`:            0,
			`bri`:            0,
		},
		`long_on`: {
			`on`:             true,
			`transitiontime`: 3000,
			`sat`:            0,
			`bri`:            254,
		},
	}
)

// Data stores data fo hub
type Data struct {
	Groups    map[string]*Group
	Scenes    map[string]*Scene
	Schedules map[string]*Schedule
	States    map[string]map[string]interface{}
}

// App stores informations and secret of API
type App struct {
	hub       provider.Hub
	groups    map[string]*Group
	scenes    map[string]*Scene
	schedules map[string]*Schedule
}

// NewApp creates new App from Flags' config
func NewApp() *App {
	return &App{}
}

func (a *App) sendWorkerMessage(w http.ResponseWriter, r *http.Request, payload interface{}, typeName, successMessage string) {
	message := &provider.WorkerMessage{
		ID:      utils.ShaFingerprint(payload),
		Source:  HueSource,
		Type:    typeName,
		Payload: payload,
	}

	if !a.hub.SendToWorker(message) {
		a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{
			Level:   `error`,
			Content: fmt.Sprintf(`[hue] Error while sending message %s to Worker`, typeName),
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
				a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[hue] Error while marshalling schedule config: %v`, err)})
				return
			}

			a.sendWorkerMessage(w, r, payload, `schedule/create`, fmt.Sprintf(`%s schedule has been created`, config.Name))
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
				a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[hue] Error while marshalling schedule: %v`, err)})
				return
			}

			a.sendWorkerMessage(w, r, payload, `schedule/update`, fmt.Sprintf(`%s schedule has been %s`, r.FormValue(`name`), schedule.Status))
			return
		}

		if postMethod == http.MethodDelete {
			a.sendWorkerMessage(w, r, []byte(id), `schedule/delete`, fmt.Sprintf(`%s schedule has been deleted`, r.FormValue(`name`)))
			return
		}
	}

	a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `[hue] Unknown schedule command`})
}

func (a *App) handleGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		postMethod := r.FormValue(`method`)

		if postMethod == http.MethodPatch {
			group := strings.Trim(strings.TrimPrefix(r.URL.Path, groupsRequest), `/`)
			state := r.FormValue(`state`)

			groupObj, ok := a.groups[group]
			if !ok {
				a.hub.RenderDashboard(w, r, http.StatusNotFound, &provider.Message{Level: `error`, Content: `[hue] Unknown group`})
			}

			a.sendWorkerMessage(w, r, fmt.Sprintf(`%s|%s`, group, state), `state/update`, fmt.Sprintf(`%s is now %s`, groupObj.Name, state))
			return
		}
	}

	a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `[hue] Unknown group command`})
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

		a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `[hue] Unknown command`})
	})
}

// SetHub receive Hub during init of it
func (a *App) SetHub(hub provider.Hub) {
	a.hub = hub
}

// GetWorkerSource get source of message in websocket
func (a *App) GetWorkerSource() string {
	return HueSource
}

// GetData return data for Dashboard rendering
func (a *App) GetData() interface{} {
	return &Data{
		Groups:    a.groups,
		Scenes:    a.scenes,
		Schedules: a.schedules,
		States:    States,
	}
}
