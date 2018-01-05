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
func NewApp(config map[string]*string, hueApp WorkerApp) *App {
	return &App{
		websocketURL: *config[`websocketURL`],
		secretKey:    *config[`secretKey`],
		hueApp:       hueApp,
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
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
	hueConfig := hue_worker.Flags(``)
	flag.Parse()

	hueApp := hue_worker.NewApp(hueConfig)
	app := NewApp(workerConfig, hueApp)

	app.connect()
}
