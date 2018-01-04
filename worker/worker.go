package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/hue"
	"github.com/ViBiOh/iot/provider"
	hue_worker "github.com/ViBiOh/iot/worker/hue"
	"github.com/gorilla/websocket"
)

const pingDelay = 60 * time.Second

// App stores informations and secret of API
type App struct {
	bridgeURL    string
	websocketURL string
	secretKey    string
	done         chan struct{}
	wsConn       *websocket.Conn
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		bridgeURL:    hue_worker.GetURL(*config[`bridgeIP`], *config[`username`]),
		websocketURL: *config[`websocketURL`],
		secretKey:    *config[`secretKey`],
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`bridgeIP`:     flag.String(tools.ToCamel(prefix+`bridgeIP`), ``, `[hue] IP of Bridge`),
		`username`:     flag.String(tools.ToCamel(prefix+`username`), ``, `[hue] Username for Bridge`),
		`websocketURL`: flag.String(tools.ToCamel(prefix+`websocket`), ``, `WebSocket URL`),
		`secretKey`:    flag.String(tools.ToCamel(prefix+`secretKey`), ``, `Secret Key`),
	}
}

func (a *App) writeTextMessage(content []byte) bool {
	if err := a.wsConn.WriteMessage(websocket.TextMessage, content); err != nil {
		log.Printf(`Error while writing text message %s: %v`, content, err)
		return false
	}

	return true
}

func (a *App) logger() {
	if a.writeTextMessage([]byte(a.secretKey)) {
		close(a.done)
	}
}

func (a *App) pinger() {
	for {
		time.Sleep(pingDelay)

		select {
		case <-a.done:
			return
		default:
			if err := a.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf(`Error while sending ping to websocket: %v`, err)

				close(a.done)
				return
			}
		}
	}
}

func (a *App) connect() {
	localIP, err := tools.GetLocalIP()
	if err != nil {
		log.Printf(`Error while retrieving local ips: %v`, err)
		return
	}

	headers := http.Header{}
	headers.Set(httputils.ForwardedForHeader, localIP.String())

	ws, _, err := websocket.DefaultDialer.Dial(a.websocketURL, headers)
	if ws != nil {
		defer ws.Close()
	}
	if err != nil {
		log.Printf(`Error while dialing to websocket %s: %v`, a.websocketURL, err)
		return
	}

	a.wsConn = ws
	a.done = make(chan struct{})
	log.Print(`Connection established`)

	a.logger()
	go a.pinger()

	input := make(chan []byte)

	go func() {
		for {
			messageType, p, err := ws.ReadMessage()
			if messageType == websocket.CloseMessage {
				close(a.done)
				return
			}

			if err != nil {
				log.Printf(`Error while reading from websocket: %v`, err)
				close(a.done)
				return
			}

			if messageType == websocket.TextMessage {
				if bytes.HasPrefix(p, provider.ErrorPrefix) {
					log.Printf(`Error received from API: %s`, bytes.TrimPrefix(p, provider.ErrorPrefix))
				} else {
					input <- p
				}
			}
		}
	}()

	for {
		select {
		case <-a.done:
			close(input)
			return
		case msg := <-input:
			if bytes.Equal(msg, hue.StatusRequest) {
				if lights, err := hue_worker.ListLightsJSON(a.bridgeURL); err != nil && !provider.WriteErrorMessage(ws, err) {
					close(a.done)
				} else if !a.writeTextMessage(append(hue.LightsPrefix, lights...)) {
					close(a.done)
				}
			} else if state, ok := hue.States[string(msg)]; ok {
				hue_worker.UpdateAllState(a.bridgeURL, state)
			}
		}
	}
}

func main() {
	workerConfig := Flags(``)
	flag.Parse()

	app := NewApp(workerConfig)
	app.connect()
}
