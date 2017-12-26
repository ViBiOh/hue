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

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`secretKey`: flag.String(tools.ToCamel(prefix+`SecretKey`), ``, `[hue] Secret Key between worker and API`),
	}
}

// NewWebsocketHandler create Websockethandler from Flags' config
func NewWebsocketHandler(config map[string]*string) http.Handler {
	secretKey := *config[`secretKey`]

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
			return
		}
		if messageType != websocket.TextMessage {
			ws.WriteMessage(websocket.TextMessage, []byte(`First message should be a Text Message`))
			return
		}
		if string(p) != secretKey {
			ws.WriteMessage(websocket.TextMessage, []byte(`First message should be the Secret Key`))
			return
		}

		for {
			messageType, p, err := ws.ReadMessage()
			if messageType == websocket.CloseMessage {
				return
			}

			if err != nil {
				log.Print(err)
				return
			}

			if err = ws.WriteMessage(messageType, p); err != nil {
				log.Print(err)
				return
			}
		}
	})
}

// NewHandler create Handler from Flags' config
func NewHandler(config map[string]*string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httputils.NotFound(w)
	})
}
