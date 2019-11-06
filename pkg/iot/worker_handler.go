package iot

import (
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/iot/pkg/provider"
)

func (a *App) handleTextMessage(p []byte) {
	var workerMessage provider.WorkerMessage
	if err := json.Unmarshal(p, &workerMessage); err != nil {
		logger.Error("%s", err)
		return
	}

	if outputChan, ok := a.workerCalls.Load(workerMessage.ID); ok {
		outputChan.(chan *provider.WorkerMessage) <- &workerMessage
	}

	if workerProvider, ok := a.workerProviders[workerMessage.Source]; ok {
		if err := workerProvider.WorkerHandler(&workerMessage); err != nil {
			logger.Error("%s", err)
		}

		return
	}

	logger.Error("%s", fmt.Errorf("no provider found for message: %#v", workerMessage))
}

// HandleWorker listen from worker
func (a *App) HandleWorker() {
	logger.Info("Connecting to MQTT %s", a.subscribeTopic)
	err := a.mqttClient.Subscribe(a.subscribeTopic, a.handleTextMessage)
	if err != nil {
		logger.Error("%s", err)
	}
}
