package hue

import (
	"bytes"
	"encoding/json"
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

// LightsPrefix prefix for passing lights status
var LightsPrefix = []byte(`lights `)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

// WebsocketHandler create Websockethandler
func (a *App) WebsocketHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if ws != nil {
			if a.wsConnexion == ws {
				a.wsConnexion = nil
			}

			log.Print(`WebSocket connection ended`)
			defer ws.Close()
		}
		if err != nil {
			log.Printf(`Error while upgrading connection: %v`, err)
			return
		}

		messageType, p, err := ws.ReadMessage()
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte(fmt.Errorf(`Error while reading first message: %v`, err).Error()))
		} else if messageType != websocket.TextMessage {
			ws.WriteMessage(websocket.TextMessage, []byte(`First message should be a Text Message`))
		} else if string(p) != a.secretKey {
			ws.WriteMessage(websocket.TextMessage, []byte(`First message should be the Secret Key`))
			return
		}

		log.Printf(`New websocket connexion setted up from %s`, httputils.GetIP(r))
		if a.wsConnexion != nil {
			a.wsConnexion.Close()
		}
		a.wsConnexion = ws

		ws.WriteMessage(websocket.TextMessage, []byte(`status`))

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
				if bytes.HasPrefix(p, []byte(LightsPrefix)) {
					var lights []Light

					if err := json.Unmarshal(bytes.TrimPrefix(LightsPrefix, p), &lights); err != nil {
						log.Printf(`Error while unmarshalling lights: %v`, err)
					} else {
						a.lights = lights
					}
				}
			}
		}
	})
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.wsConnexion != nil {
			event := strings.TrimPrefix(r.URL.Path, `/`)
			if err := a.wsConnexion.WriteMessage(websocket.TextMessage, []byte(event)); err != nil {
				a.renderer.RenderDashboard(w, r, http.StatusInternalServerError, &provider.Message{Level: `error`, Content: fmt.Sprintf(`Error while talking to Worker: %v`, err)})
			} else {
				a.wsConnexion.WriteMessage(websocket.TextMessage, []byte(`status`))
				a.renderer.RenderDashboard(w, r, http.StatusOK, &provider.Message{Level: `success`, Content: fmt.Sprintf(`Lights turned to %s`, event)})
			}
		} else {
			a.renderer.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: `Worker is not listening`})
		}
	})
}

// SetRenderer handle store of Renderer
func (a *App) SetRenderer(r provider.Renderer) {
	a.renderer = r
}

// GetData return data provided to renderer
func (a *App) GetData() interface{} {
	on := 0
	for _, light := range a.lights {
		if light.State.On {
			on++
		}
	}

	return fmt.Sprintf(`%d / %d`, on, len(a.lights))
}
