package hue

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/provider"
	"github.com/gorilla/websocket"
)

var (
	// LightsPrefix prefix for passing lights status
	LightsPrefix = []byte(`lights `)

	// StatusRequest payload
	StatusRequest = []byte(`status`)

	// States available states of lights
	States = map[string]string{
		`off`:    `{"on":false,"transitiontime":30}`,
		`dimmed`: `{"on":true,"transitiontime":30,"sat":0,"bri":0}`,
		`bright`: `{"on":true,"transitiontime":30,"sat":0,"bri":254}`,
	}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Data stores data fo renderer
type Data struct {
	Online bool
	Status string
}

// Light description
type Light struct {
	Name  string
	State struct {
		On bool
	}
}

// App stores informations and secret of API
type App struct {
	secretKey   string
	wsConnexion *websocket.Conn
	renderer    provider.Renderer
	lights      []Light
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

// WebsocketHandler create Websockethandler
func (a *App) WebsocketHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if ws != nil {
			if a.wsConnexion == ws {
				a.wsConnexion = nil
			}

			defer ws.Close()
		}
		if err != nil {
			log.Printf(`Error while upgrading connection: %v`, err)
			return
		}

		if !a.checkWorker(ws) {
			return
		}

		log.Printf(`Worker connection from %s and %s`, httputils.GetIP(r), ws.RemoteAddr())
		if a.wsConnexion != nil {
			a.wsConnexion.Close()
		}
		a.wsConnexion = ws

		ws.WriteMessage(websocket.TextMessage, StatusRequest)

		for {
			messageType, p, err := ws.ReadMessage()
			if messageType == websocket.CloseMessage {
				return
			}

			if err != nil {
				log.Print(err)
				return
			}

			if messageType == websocket.TextMessage {
				if bytes.HasPrefix(p, LightsPrefix) {
					var lights []Light
					jsonData := bytes.TrimPrefix(p, LightsPrefix)

					if err := json.Unmarshal(jsonData, &lights); err != nil {
						log.Printf(`Error while unmarshalling lights "%s": %v`, jsonData, err)
					} else {
						a.lights = lights
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
		} else if a.wsConnexion == nil {
			a.renderer.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `Worker is not listening`})
		} else {
			event := strings.TrimPrefix(r.URL.Path, `/`)
			if err := a.wsConnexion.WriteMessage(websocket.TextMessage, []byte(event)); err != nil {
				a.renderer.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: fmt.Sprintf(`Error while talking to Worker: %v`, err)})
			} else {
				a.wsConnexion.WriteMessage(websocket.TextMessage, StatusRequest)
				a.renderer.RenderDashboard(w, r, http.StatusOK, &provider.Message{Level: `success`, Content: fmt.Sprintf(`Lights turned to %s`, event)})
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
	data := &Data{
		Online: a.wsConnexion != nil,
		Status: ``,
	}

	if len(a.lights) == 0 {
		return data
	}

	on := 0
	for _, light := range a.lights {
		if light.State.On {
			on++
		}
	}

	data.Status = fmt.Sprintf(`%d / %d`, on, len(a.lights))
	return data
}
