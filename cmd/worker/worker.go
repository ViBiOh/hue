package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/hue"
	"github.com/ViBiOh/iot/pkg/provider"
	hue_worker "github.com/ViBiOh/iot/pkg/worker/hue"
	"github.com/gorilla/websocket"
)

const (
	pingID    = `ping`
	pingDelay = 60 * time.Second
)

// WorkerApp app that plugs to worker
type WorkerApp interface {
	Handle(*provider.WorkerMessage) (*provider.WorkerMessage, error)
	Ping() ([]*provider.WorkerMessage, error)
}

// App stores informations and secret of API
type App struct {
	websocketURL string
	secretKey    string
	hueApp       WorkerApp
	done         chan struct{}
	wsConn       *websocket.Conn
	debug        bool
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}, hueApp WorkerApp) *App {
	return &App{
		websocketURL: *config[`websocketURL`].(*string),
		secretKey:    *config[`secretKey`].(*string),
		hueApp:       hueApp,
		debug:        *config[`debug`].(*bool),
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`websocketURL`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `websocket`)), ``, `WebSocket URL`),
		`secretKey`:    flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `secretKey`)), ``, `Secret Key`),
		`debug`:        flag.Bool(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Debug`)), false, `Enable debug`),
	}
}

func (a *App) auth() {
	if err := a.wsConn.WriteMessage(websocket.TextMessage, []byte(a.secretKey)); err != nil {
		log.Printf(`Error while sending auth message: %v`, err)
		close(a.done)
	}
}

func (a *App) pinger() {
	for {
		select {
		case <-a.done:
			return
		default:
			messages, err := a.hueApp.Ping()

			if err != nil {
				if err := provider.WriteErrorMessage(a.wsConn, hue.HueSource, err); err != nil {
					log.Print(err)
					close(a.done)
				}
			} else {
				for _, message := range messages {
					if err := provider.WriteMessage(a.wsConn, message); err != nil {
						log.Print(err)
						close(a.done)
					}
				}
			}
		}

		time.Sleep(pingDelay)
	}
}

func (a *App) connect() {
	ws, _, err := websocket.DefaultDialer.Dial(a.websocketURL, nil)
	if ws != nil {
		defer func() {
			if err := ws.Close(); err != nil {
				log.Printf(`Error while closing websocket connection: %v`, err)
			}
		}()
	}
	if err != nil {
		log.Printf(`Error while dialing to websocket %s: %v`, a.websocketURL, err)
		return
	}

	a.wsConn = ws
	a.done = make(chan struct{})
	log.Print(`Websocket connection established`)

	a.auth()
	go a.pinger()

	input := make(chan *provider.WorkerMessage)

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
				var workerMessage provider.WorkerMessage
				if err := json.Unmarshal(p, &workerMessage); err != nil {
					log.Printf(`Error while unmarshalling worker message: %v`, err)
				} else {
					input <- &workerMessage
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
			if a.debug {
				log.Printf(`[%s] %s: %s`, p.Source, p.Type, p.Payload)
			}

			if p.Source == hue.HueSource {
				output, err := a.hueApp.Handle(p)

				if err != nil {
					if err := provider.WriteErrorMessage(a.wsConn, hue.HueSource, err); err != nil {
						log.Print(err)
						close(a.done)
					}
				} else if output != nil {
					if err := provider.WriteMessage(a.wsConn, output); err != nil {
						log.Print(err)
						close(a.done)
					}
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

	hueApp, err := hue_worker.NewApp(hueConfig, *workerConfig[`debug`].(*bool))
	if err != nil {
		log.Fatalf(`Error while creating hue app: %s`, err)
	}

	app := NewApp(workerConfig, hueApp)

	app.connect()
}