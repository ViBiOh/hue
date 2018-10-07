package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/tools"
	hue_worker "github.com/ViBiOh/iot/pkg/hue/worker"
	netatmo_worker "github.com/ViBiOh/iot/pkg/netatmo/worker"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/gorilla/websocket"
)

const (
	pingDelay = 60 * time.Second
)

// App stores informations and secret of API
type App struct {
	websocketURL string
	secretKey    string
	workers      map[string]provider.Worker
	done         chan struct{}
	wsConn       *websocket.Conn
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string, workers []provider.Worker) *App {
	workersMap := make(map[string]provider.Worker, len(workers))
	for _, worker := range workers {
		workersMap[worker.GetSource()] = worker
	}

	return &App{
		websocketURL: strings.TrimSpace(*config[`websocketURL`]),
		secretKey:    strings.TrimSpace(*config[`secretKey`]),
		workers:      workersMap,
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`websocketURL`: flag.String(tools.ToCamel(fmt.Sprintf(`%sWebsocket`, prefix)), ``, `WebSocket URL`),
		`secretKey`:    flag.String(tools.ToCamel(fmt.Sprintf(`%sSecretKey`, prefix)), ``, `Secret Key`),
	}
}

func (a *App) auth() {
	if err := a.wsConn.WriteMessage(websocket.TextMessage, []byte(a.secretKey)); err != nil {
		rollbar.LogError(`error while sending auth message: %v`, err)
		close(a.done)
	}
}

func (a *App) pingWorkers() {
	ctx := context.Background()
	workersCount := len(a.workers)

	inputs, results, errors := tools.ConcurrentAction(uint(workersCount), func(e interface{}) (interface{}, error) {
		if worker, ok := e.(provider.Worker); ok {
			return worker.Ping(ctx)
		}

		return nil, fmt.Errorf(`unrecognized worker type: %+v`, e)
	})

	go func() {
		defer close(inputs)

		for _, worker := range a.workers {
			inputs <- worker
		}
	}()

	for i := 0; i < workersCount; i++ {
		select {
		case err := <-errors:
			source := `unknown`
			if worker, ok := err.Input.(provider.Worker); ok {
				source = worker.GetSource()
			}

			if err := provider.WriteErrorMessage(a.wsConn, source, err.Err); err != nil {
				rollbar.LogError(`%v`, err)
			}
			break

		case result := <-results:
			if messages, ok := result.([]*provider.WorkerMessage); ok {
				for _, message := range messages {
					if err := provider.WriteMessage(context.Background(), a.wsConn, message); err != nil {
						rollbar.LogError(`%v`, err)
					}
				}
			} else {
				rollbar.LogError(`unrecognized message type: %+v`, result)
			}
			break
		}
	}
}

func (a *App) pinger() {
	for {
		select {
		case <-a.done:
			return
		default:
			a.pingWorkers()
		}

		time.Sleep(pingDelay)
	}
}

func (a *App) handleMessage(p *provider.WorkerMessage) {
	ctx, span, err := opentracing.ExtractSpanFromMap(context.Background(), p.Tracing, p.Action)
	if err != nil {
		rollbar.LogError(`%v`, err)
	}
	if span != nil {
		defer span.Finish()
	}

	if worker, ok := a.workers[p.Source]; ok {
		output, err := worker.Handle(ctx, p)

		if err != nil {
			if err := provider.WriteErrorMessage(a.wsConn, p.Source, err); err != nil {
				rollbar.LogError(`%v`, err)
			}
		} else if output != nil {
			if err := provider.WriteMessage(ctx, a.wsConn, output); err != nil {
				rollbar.LogError(`%v`, err)
			}
		}
	} else {
		rollbar.LogError(`Unknown request: %s`, p)
	}
}

func (a *App) connect() {
	ws, _, err := websocket.DefaultDialer.Dial(a.websocketURL, nil)
	if ws != nil {
		defer func() {
			if err := ws.Close(); err != nil {
				rollbar.LogError(`Error while closing websocket connection: %v`, err)
			}
		}()
	}
	if err != nil {
		rollbar.LogError(`Error while dialing to websocket %s: %v`, a.websocketURL, err)
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
				rollbar.LogError(`Error while reading from websocket: %v`, err)
				close(a.done)
				return
			}

			if messageType == websocket.TextMessage {
				var workerMessage provider.WorkerMessage
				if err := json.Unmarshal(p, &workerMessage); err != nil {
					rollbar.LogError(`Error while unmarshalling worker message: %v`, err)
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
			a.handleMessage(p)
		}
	}
}

func main() {
	workerConfig := Flags(``)
	hueConfig := hue_worker.Flags(`hue`)
	netatmoConfig := netatmo_worker.Flags(`netatmo`)
	opentracingConfig := opentracing.Flags(`tracing`)
	rollbarConfig := rollbar.Flags(`rollbar`)
	flag.Parse()

	opentracing.NewApp(opentracingConfig)
	rollbar.NewApp(rollbarConfig)

	hueApp, err := hue_worker.NewApp(hueConfig)
	if err != nil {
		rollbar.LogError(`Error while creating hue app: %s`, err)
		os.Exit(1)
	}
	netatmoApp := netatmo_worker.NewApp(netatmoConfig)
	app := NewApp(workerConfig, []provider.Worker{hueApp, netatmoApp})

	app.connect()
}
