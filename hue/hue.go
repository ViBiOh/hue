package hue

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/iot"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// App stores informations and secret of API
type App struct {
	secretKey   string
	wsConnexion *websocket.Conn
	iotApp      *iot.App
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string, iotApp *iot.App) *App {
	return &App{
		secretKey: *config[`secretKey`],
		iotApp:    iotApp,
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
				log.Printf(`%s`, p)
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
				a.iotApp.RenderDashboard(w, r, http.StatusInternalServerError, &iot.Message{Level: `error`, Content: fmt.Sprintf(`Error while requesting Hue Worker: %v`, err)})
			} else {
				a.iotApp.RenderDashboard(w, r, http.StatusInternalServerError, &iot.Message{Level: `success`, Content: fmt.Sprintf(`Lights turned to %s`, event)})
			}
		} else {
			a.iotApp.RenderDashboard(w, r, http.StatusServiceUnavailable, &iot.Message{Level: `error`, Content: `Hue Worker is not listening`})
		}
	})
}
