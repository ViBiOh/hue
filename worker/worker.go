package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/request"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/hue"
	"github.com/ViBiOh/iot/provider"
	hue_worker "github.com/ViBiOh/iot/worker/hue"
	"github.com/gorilla/websocket"
)

const pingDelay = 60 * time.Second

// WorkerApp app that plugs to worker
type WorkerApp interface {
	Handle([]byte) ([]byte, error)
}

// App stores informations and secret of API
type App struct {
	websocketURL string
	secretKey    string
	hueApp       WorkerApp
	done         chan struct{}
	wsConn       *websocket.Conn
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}, hueApp WorkerApp) *App {
	return &App{
		websocketURL: *config[`websocketURL`].(*string),
		secretKey:    *config[`secretKey`].(*string),
		hueApp:       hueApp,
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`websocketURL`: flag.String(tools.ToCamel(prefix+`websocket`), ``, `WebSocket URL`),
		`secretKey`:    flag.String(tools.ToCamel(prefix+`secretKey`), ``, `Secret Key`),
	}
}

func (a *App) auth() {
	if !provider.WriteTextMessage(a.wsConn, []byte(a.secretKey)) {
		close(a.done)
	}
}

func (a *App) pinger() {
	for {
		select {
		case <-a.done:
			return
		default:
			var output []byte
			var err error

			output, err = a.hueApp.Handle(hue.GroupsPrefix)
			if err != nil && !provider.WriteErrorMessage(a.wsConn, err) {
				close(a.done)
			} else if output != nil && !provider.WriteTextMessage(a.wsConn, append(hue.WebSocketPrefix, output...)) {
				close(a.done)
			}

			output, err = a.hueApp.Handle(hue.SchedulesPrefix)
			if err != nil && !provider.WriteErrorMessage(a.wsConn, err) {
				close(a.done)
			} else if output != nil && !provider.WriteTextMessage(a.wsConn, append(hue.WebSocketPrefix, output...)) {
				close(a.done)
			}

			output, err = a.hueApp.Handle(hue.ScenesPrefix)
			if err != nil && !provider.WriteErrorMessage(a.wsConn, err) {
				close(a.done)
			} else if output != nil && !provider.WriteTextMessage(a.wsConn, append(hue.WebSocketPrefix, output...)) {
				close(a.done)
			}
		}

		time.Sleep(pingDelay)
	}
}

func (a *App) connect() {
	localIP, err := tools.GetLocalIP()
	if err != nil {
		log.Printf(`Error while retrieving local ips: %v`, err)
		return
	}

	headers := http.Header{}
	headers.Set(request.ForwardedForHeader, localIP.String())

	ws, _, err := websocket.DefaultDialer.Dial(a.websocketURL, headers)
	if ws != nil {
		defer func() {
			if err := ws.Close(); err != nil {
				log.Printf(`Error while closing connection: %v`, err)
			}
		}()
	}
	if err != nil {
		log.Printf(`Error while dialing to websocket %s: %v`, a.websocketURL, err)
		return
	}

	a.wsConn = ws
	a.done = make(chan struct{})
	log.Print(`Connection established`)

	a.auth()
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
		case p := <-input:
			if bytes.HasPrefix(p, hue.WebSocketPrefix) {
				output, err := a.hueApp.Handle(bytes.TrimPrefix(p, hue.WebSocketPrefix))

				if err != nil && !provider.WriteErrorMessage(a.wsConn, err) {
					close(a.done)
				} else if output != nil && !provider.WriteTextMessage(a.wsConn, append(hue.WebSocketPrefix, output...)) {
					close(a.done)
				}
			} else {
				log.Printf(`Unknown request: %s`, p)
			}
		}
	}
}

func main() {
	workerConfig := Flags(``)
	hueConfig := hue_worker.Flags(`hue`)
	flag.Parse()

	hueApp, err := hue_worker.NewApp(hueConfig)
	if err != nil {
		log.Fatalf(`Error while creating hue app: %s`, err)
	}

	app := NewApp(workerConfig, hueApp)

	app.connect()
}
