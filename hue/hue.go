package hue

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
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

func handleRedirect(w http.ResponseWriter, r *http.Request, event string, err error) {
	if err != nil {
		log.Printf(`Error while querying Hue WebSocket: %v`, err)
		http.Redirect(w, r, fmt.Sprintf(`/?message_level=%s&message_content=%s`, `error`, `Error while requesting Hue`), http.StatusFound)
	} else {
		http.Redirect(w, r, fmt.Sprintf(`/?message_level=%s&message_content=%s`, `success`, fmt.Sprintf(`Lights turned %s`, event)), http.StatusFound)
	}
}

// Handler create Handler from Flags' config
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.wsConnexion != nil {
			if r.URL.Path == `/on` {
				handleRedirect(w, r, `on`, a.wsConnexion.WriteMessage(websocket.TextMessage, []byte(`bright`)))
			} else if r.URL.Path == `/off` {
				handleRedirect(w, r, `off`, a.wsConnexion.WriteMessage(websocket.TextMessage, []byte(`off`)))
			} else {
				httputils.NotFound(w)
			}
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})
}
