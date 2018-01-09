package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/iot/provider"
)

var (
	// WebSocketPrefix ws message prefix for all hue commands
	WebSocketPrefix = []byte(`hue `)

	// GroupsPrefix ws message prefix for groups command
	GroupsPrefix = []byte(`groups `)

	// StatePrefix ws message prefix for state command
	StatePrefix = []byte(`state `)

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
	}
)

// Group description
type Group struct {
	Name   string
	On     bool
	OnOff  bool
	Lights []string
	State  struct {
		On bool
	}
}

// Light description
type Light struct {
	Type  string
	State struct {
		On bool
	}
}

// Data stores data fo hub
type Data struct {
	Groups map[string]*Group
}

// App stores informations and secret of API
type App struct {
	hub    provider.Hub
	groups map[string]*Group
}

// NewApp creates new App from Flags' config
func NewApp() *App {
	return &App{}
}

func (a *App) sendToWorker(payload []byte) bool {
	return a.hub.SendToWorker(append(WebSocketPrefix, payload...))
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `Unknown command`})
		} else if r.URL.Path == `/state` {
			params := r.URL.Query()

			group := params.Get(`group`)
			state := params.Get(`value`)

			if !a.sendToWorker(append(StatePrefix, []byte(fmt.Sprintf(`%s|%s`, group, state))...)) {
				a.hub.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: `Error while sending command to Worker`})
			} else {
				a.hub.RenderDashboard(w, r, http.StatusOK, &provider.Message{Level: `success`, Content: fmt.Sprintf(`%s is now %s`, a.groups[group].Name, state)})
			}
		}
	})
}

// SetHub receive Hub during init of it
func (a *App) SetHub(hub provider.Hub) {
	a.hub = hub
}

// GetWorkerPrefix get prefix of message in websocket
func (a *App) GetWorkerPrefix() []byte {
	return WebSocketPrefix
}

// GetData return data for Dashboard rendering
func (a *App) GetData() interface{} {
	return &Data{
		Groups: a.groups,
	}
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(payload []byte) {
	if bytes.HasPrefix(payload, GroupsPrefix) {
		if err := json.Unmarshal(bytes.TrimPrefix(payload, GroupsPrefix), &a.groups); err != nil {
			log.Printf(`[hue] Error while unmarshalling groups: %v`, err)
		}
	} else {
		log.Printf(`[hue] Unknown command`)
	}
}
