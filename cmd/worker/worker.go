package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/tools"
	hue_worker "github.com/ViBiOh/iot/pkg/hue/worker"
	netatmo_worker "github.com/ViBiOh/iot/pkg/netatmo/worker"
	"github.com/ViBiOh/iot/pkg/provider"
	sonos_worker "github.com/ViBiOh/iot/pkg/sonos/worker"
	"github.com/gorilla/websocket"
)

const (
	pingDelay = 60 * time.Second
)

// Config of package
type Config struct {
	websocketURL *string
	secretKey    *string
}

// App of package
type App struct {
	websocketURL string
	secretKey    string
	workers      map[string]provider.Worker
	done         chan struct{}
	wsConn       *websocket.Conn
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		websocketURL: fs.String(tools.ToCamel(fmt.Sprintf(`%sWebsocket`, prefix)), ``, `WebSocket URL`),
		secretKey:    fs.String(tools.ToCamel(fmt.Sprintf(`%sSecretKey`, prefix)), ``, `Secret Key`),
	}
}

// New creates new App from Config
func New(config Config, workers []provider.Worker) *App {
	workersMap := make(map[string]provider.Worker, len(workers))
	for _, worker := range workers {
		workersMap[worker.GetSource()] = worker
	}

	return &App{
		websocketURL: strings.TrimSpace(*config.websocketURL),
		secretKey:    strings.TrimSpace(*config.secretKey),
		workers:      workersMap,
	}
}

func (a *App) auth() {
	if err := a.wsConn.WriteMessage(websocket.TextMessage, []byte(a.secretKey)); err != nil {
		logger.Error(`%+v`, errors.WithStack(err))
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

		return nil, errors.New(`unrecognized worker type: %+v`, e)
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
				logger.Error(`%+v`, err)
			}
			break

		case result := <-results:
			if messages, ok := result.([]*provider.WorkerMessage); ok {
				for _, message := range messages {
					if err := provider.WriteMessage(ctx, a.wsConn, message); err != nil {
						logger.Error(`%+v`, err)
					}
				}
			} else {
				logger.Error(`unrecognized message type: %+v`, result)
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
		logger.Error(`%+v`, errors.WithStack(err))
	}
	if span != nil {
		defer span.Finish()
	}

	if worker, ok := a.workers[p.Source]; ok {
		output, err := worker.Handle(ctx, p)

		if err != nil {
			logger.Error(`%+v`, err)

			if err := provider.WriteErrorMessage(a.wsConn, p.Source, err); err != nil {
				logger.Error(`%+v`, err)
			}
		}

		if output != nil {
			if err := provider.WriteMessage(ctx, a.wsConn, output); err != nil {
				logger.Error(`%+v`, err)
			}
		}

		return
	}

	logger.Error(`unknown request: %s`, p)
}

func (a *App) connect() {
	ws, _, err := websocket.DefaultDialer.Dial(a.websocketURL, nil)
	if ws != nil {
		defer func() {
			if err := ws.Close(); err != nil {
				logger.Error(`%+v`, errors.WithStack(err))
			}
		}()
	}
	if err != nil {
		logger.Error(`%+v`, errors.WithStack(err))
		return
	}

	a.wsConn = ws
	a.done = make(chan struct{})
	logger.Info(`Websocket connection established`)

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
				logger.Error(`%+v`, errors.WithStack(err))
				close(a.done)
				return
			}

			if messageType == websocket.TextMessage {
				var workerMessage provider.WorkerMessage
				if err := json.Unmarshal(p, &workerMessage); err != nil {
					logger.Error(`%+v`, errors.WithStack(err))
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
	fs := flag.NewFlagSet(`iot-worker`, flag.ExitOnError)

	workerConfig := Flags(fs, ``)
	hueConfig := hue_worker.Flags(fs, `hue`)
	netatmoConfig := netatmo_worker.Flags(fs, `netatmo`)
	sonosConfig := sonos_worker.Flags(fs, `sonos`)

	if err := fs.Parse(os.Args[1:]); err != nil {
		logger.Fatal(`%+v`, err)
	}

	hueApp, err := hue_worker.New(hueConfig)
	if err != nil {
		logger.Error(`%+v`, err)
		os.Exit(1)
	}
	netatmoApp := netatmo_worker.New(netatmoConfig)
	sonosApp := sonos_worker.New(sonosConfig)
	app := New(workerConfig, []provider.Worker{hueApp, netatmoApp, sonosApp})

	app.connect()
}
