package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/provider"
	"github.com/gorilla/websocket"
)

var (
	// WebSocketPrefix ws message prefix for all hue commands
	WebSocketPrefix = []byte(`hue `)

	// GroupsPrefix ws message prefix for groups command
	GroupsPrefix = []byte(`groups `)

	// StatePrefix ws message prefix for state command
	StatePrefix = []byte(`state `)

	// States available states of lights
	States = map[string]string{
		`off`:    `{"on":false,"transitiontime":30}`,
		`on`:     `{"on":true,"transitiontime":30,"sat":0,"bri":254}`,
		`dimmed`: `{"on":true,"transitiontime":30,"sat":0,"bri":0}`,
	}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Group description
type Group struct {
	Name   string
	OnOff  bool
	Lights []string
	State  struct {
		On bool
	}
}

// Light description
type Light struct {
	Type string
}

// Data stores data fo renderer
type Data struct {
	Online bool
	Groups map[string]*Group
}

// App stores informations and secret of API
type App struct {
	secretKey   string
	heaterGroup string
	wsConn      *websocket.Conn
	renderer    provider.Renderer
	groups      map[string]*Group
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		secretKey: *config[`secretKey`],
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`secretKey`: flag.String(tools.ToCamel(prefix+`SecretKey`), ``, `[hue] Secret Key between worker and API`),
	}
}

func (a *App) checkWorker(ws *websocket.Conn) bool {
	messageType, p, err := ws.ReadMessage()
	if err != nil {
		provider.WriteErrorMessage(ws, fmt.Errorf(`Error while reading first message: %v`, err))
		return false
	}
	if messageType != websocket.TextMessage {
		provider.WriteErrorMessage(ws, errors.New(`First message should be a Text Message`))
		return false
	}
	if string(p) != a.secretKey {
		provider.WriteErrorMessage(ws, errors.New(`First message should be the Secret Key`))
		return false
	}

	return true
}

func (a *App) writeWorker(content []byte) bool {
	return provider.WriteTextMessage(a.wsConn, append(WebSocketPrefix, content...))
}

// WebsocketHandler create Websockethandler
func (a *App) WebsocketHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if ws != nil {
			defer func() {
				if a.wsConn == ws {
					a.wsConn = nil
				}

				ws.Close()
			}()
		}
		if err != nil {
			log.Printf(`Error while upgrading connection: %v`, err)
			return
		}

		if !a.checkWorker(ws) {
			return
		}

		log.Printf(`Worker connection from %s`, httputils.GetIP(r))
		if a.wsConn != nil {
			a.wsConn.Close()
		}
		a.wsConn = ws

		if !a.writeWorker(GroupsPrefix) {
			return
		}

		for {
			messageType, p, err := ws.ReadMessage()
			if messageType == websocket.CloseMessage {
				return
			}

			if err != nil {
				log.Printf(`Error while reading from websocket: %v`, err)
				return
			}

			if messageType == websocket.TextMessage {
				if bytes.HasPrefix(p, GroupsPrefix) {
					if err := json.Unmarshal(bytes.TrimPrefix(p, GroupsPrefix), &a.groups); err != nil {
						log.Printf(`Error while unmarshalling groups: %v`, err)
					}
				} else if bytes.HasPrefix(p, provider.ErrorPrefix) {
					log.Printf(`Error received from worker: %s`, bytes.TrimPrefix(p, provider.ErrorPrefix))
				}
			}
		}
	})
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			a.renderer.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `Unknown command`})
		} else if a.wsConn == nil {
			a.renderer.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `Worker is not listening`})
		} else if r.URL.Path == `/state` {
			params := r.URL.Query()

			group := params.Get(`group`)
			state := params.Get(`value`)

			if !a.writeWorker(append(StatePrefix, []byte(fmt.Sprintf(`%s|%s`, group, state))...)) {
				a.renderer.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: `Error while talking to Worker`})
			} else {
				a.renderer.RenderDashboard(w, r, http.StatusOK, &provider.Message{Level: `success`, Content: fmt.Sprintf(`%s is now %s`, a.groups[group].Name, state)})
			}
		}
	})
}

// SetRenderer handle store of Renderer
func (a *App) SetRenderer(r provider.Renderer) {
	a.renderer = r
}

// GetData return data provided to renderer
func (a *App) GetData() interface{} {
	return &Data{
		Online: a.wsConn != nil,
		Groups: a.groups,
	}
}
