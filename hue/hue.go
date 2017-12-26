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

var (
	secretKey string
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

// Init retrieves config from CLI args
func Init(config map[string]*string) error {
	secretKey = *config[`secretKey`]

	return nil
}

func workerHandler() http.Handler {
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

// Handler for Hue Request
func Handler() http.Handler {
	ws := workerHandler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/ws` {
			ws.ServeHTTP(w, r)
		} else {
			httputils.NotFound(w)
		}
	})
}
