package iot

import (
	"encoding/json"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/iot/pkg/provider"
)

func (a *App) handleTextMessage(p []byte) {
	var workerMessage provider.WorkerMessage
	if err := json.Unmarshal(p, &workerMessage); err != nil {
		logger.Error("%#v", errors.WithStack(err))
		return
	}

	if outputChan, ok := a.workerCalls.Load(workerMessage.ID); ok {
		outputChan.(chan *provider.WorkerMessage) <- &workerMessage
	}

	if workerProvider, ok := a.workerProviders[workerMessage.Source]; ok {
		if err := workerProvider.WorkerHandler(&workerMessage); err != nil {
			logger.Error("%#v", err)
		}

		return
	}

	logger.Error("%#v", errors.New("no provider found for message: %#v", workerMessage))
}

// HandleWorker listen from worker
func (a *App) HandleWorker() {
	logger.Info("Connecting to MQTT %s", a.subscribeTopic)
	err := a.mqttClient.Subscribe(a.subscribeTopic, a.handleTextMessage)
	if err != nil {
		logger.Error("%#v", err)
	}
}
