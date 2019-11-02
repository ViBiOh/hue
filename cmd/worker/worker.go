package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/concurrent"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	hue_worker "github.com/ViBiOh/iot/pkg/hue/worker"
	"github.com/ViBiOh/iot/pkg/mqtt"
	"github.com/ViBiOh/iot/pkg/provider"
	sonos_worker "github.com/ViBiOh/iot/pkg/sonos/worker"
)

const (
	pingDelay = 60 * time.Second
)

// Config of package
type Config struct {
	publish   *string
	subscribe *string
}

// App of package
type App struct {
	publishTopics  []string
	subscribeTopic string
	workers        map[string]provider.Worker
	handlers       map[string]provider.WorkerHandler
	mqttClient     *mqtt.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		publish:   flags.New(prefix, "worker").Name("Publish").Default("local,remote").Label("Topics to publish to, comma separated").ToString(fs),
		subscribe: flags.New(prefix, "worker").Name("Subscribe").Default("worker").Label("Topic to subscribe to").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, workers []provider.Worker, mqttClient *mqtt.App) *App {
	workersMap := make(map[string]provider.Worker, len(workers))
	handlersMap := make(map[string]provider.WorkerHandler, 0)

	for _, worker := range workers {
		if !worker.Enabled() {
			logger.Info("Worker %s disabled", worker.GetSource())
			continue
		}

		workersMap[worker.GetSource()] = worker

		if handler, ok := worker.(provider.WorkerHandler); ok {
			handlersMap[worker.GetSource()] = handler

			logger.Info("Registered %s handler", worker.GetSource())

			if _, ok := worker.(provider.Pinger); ok && worker.Enabled() {
				logger.Info("Ping enabled for %s", worker.GetSource())
			}
		}

		if starter, ok := worker.(provider.Starter); ok {
			logger.Info("Starting %s", worker.GetSource())
			starter.Start()
		}
	}

	return &App{
		workers:        workersMap,
		handlers:       handlersMap,
		mqttClient:     mqttClient,
		publishTopics:  strings.Split(strings.TrimSpace(*config.publish), ","),
		subscribeTopic: strings.TrimSpace(*config.subscribe),
	}
}

func (a *App) pingWorkers() {
	ctx := context.Background()
	workersCount := len(a.workers)

	action := func(e interface{}) (interface{}, error) {
		if worker, ok := e.(provider.Worker); ok {
			if !worker.Enabled() {
				return nil, nil
			}

			if pinger, ok := e.(provider.Pinger); ok {
				return pinger.Ping(ctx)
			}

			return nil, nil
		}

		return nil, fmt.Errorf("unrecognized worker type: %#v", e)
	}

	onSucces := func(output interface{}) {
		if !a.mqttClient.Enabled() {
			return
		}

		for _, message := range output.([]*provider.WorkerMessage) {
			for _, topic := range a.publishTopics {
				if err := provider.WriteMessage(ctx, a.mqttClient, topic, message); err != nil {
					logger.Error("%s", err)
				}
			}
		}
	}

	onError := func(err error) {
		logger.Error("%s", err)
	}

	inputs := concurrent.Run(uint(workersCount), action, onSucces, onError)

	for _, worker := range a.workers {
		inputs <- worker
	}
	close(inputs)
}

func (a *App) pinger() {
	for {
		a.pingWorkers()
		time.Sleep(pingDelay)
	}
}

func (a *App) handleTextMessage(p []byte) {
	var message provider.WorkerMessage
	if err := json.Unmarshal(p, &message); err != nil {
		logger.Error("%s", err)
		return
	}

	ctx := context.Background()

	if worker, ok := a.handlers[message.Source]; ok {
		output, err := worker.Handle(ctx, &message)

		if err != nil {
			logger.Error("%s", err)
		}

		if output != nil {
			if err := provider.WriteMessage(ctx, a.mqttClient, message.ResponseTo, output); err != nil {
				logger.Error("%s", err)
			}
		}

		return
	}

	logger.Error("unknown request: %s", message)
}

func (a *App) connect() {
	logger.Info("Connecting to MQTT %s", a.subscribeTopic)
	err := a.mqttClient.Subscribe(a.subscribeTopic, a.handleTextMessage)
	if err != nil {
		logger.Error("%s", err)
	}
}

func main() {
	fs := flag.NewFlagSet("iot-worker", flag.ExitOnError)

	iotConfig := Flags(fs, "")
	mqttConfig := mqtt.Flags(fs, "mqtt")
	hueConfig := hue_worker.Flags(fs, "hue")
	sonosConfig := sonos_worker.Flags(fs, "sonos")

	logger.Fatal(fs.Parse(os.Args[1:]))

	hueApp, err := hue_worker.New(hueConfig)
	if err != nil {
		logger.Error("%s", err)
		os.Exit(1)
	}

	mqttApp, err := mqtt.New(mqttConfig)
	logger.Fatal(err)

	sonosApp := sonos_worker.New(sonosConfig)

	app := New(iotConfig, []provider.Worker{hueApp, sonosApp}, mqttApp)

	if mqttApp.Enabled() {
		app.connect()
	}

	app.pinger()
}
