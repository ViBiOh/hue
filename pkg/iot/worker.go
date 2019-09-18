package iot

import (
	"context"
	"time"

	"github.com/ViBiOh/httputils/v2/pkg/logger"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	workerWaitDelay = 10 * time.Second
)

func (a *App) registerWorker(worker provider.WorkerProvider) {
	a.workerProviders[worker.GetWorkerSource()] = worker

	logger.Info("Worker registered for %s", worker.GetWorkerSource())
}

// SendToWorker sends payload to worker
func (a *App) SendToWorker(ctx context.Context, root *provider.WorkerMessage, source, action string, payload string, waitOutput bool) *provider.WorkerMessage {
	message := provider.NewWorkerMessage(root, source, action, payload)

	var outputChan chan *provider.WorkerMessage
	if waitOutput {
		outputChan = make(chan *provider.WorkerMessage)
		a.workerCalls.Store(message.ID, outputChan)

		defer a.workerCalls.Delete(message.ID)
	}

	message.ResponseTo = a.subscribeTopic

	if err := provider.WriteMessage(ctx, a.mqttClient, a.publishTopic, message); err != nil {
		return provider.NewWorkerMessage(root, message.Source, provider.WorkerErrorAction, err.Error())
	}

	if waitOutput {
		select {
		case output := <-outputChan:
			return output
		case <-time.After(workerWaitDelay):
			return provider.NewWorkerMessage(root, message.Source, provider.WorkerErrorAction, "timeout exceeded")
		}
	}

	return nil
}
