package iot

import (
	"context"
	"errors"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	workerWaitDelay = 10 * time.Second
)

var (
	// ErrWorkerResponseTimeout occurs when worker didn't respond in delay
	ErrWorkerResponseTimeout = errors.New("timeout exceeded for waiting for worker response")
)

func (a *App) registerWorker(worker provider.WorkerProvider) {
	a.workerProviders[worker.GetWorkerSource()] = worker

	logger.Info("Worker registered for %s", worker.GetWorkerSource())
}

// SendToWorker sends payload to worker
func (a *App) SendToWorker(ctx context.Context, root *provider.WorkerMessage, source, action string, payload string) (provider.WorkerMessage, error) {
	message := provider.NewWorkerMessage(root, source, action, payload)
	message.ResponseTo = a.subscribeTopic

	outputChan := make(chan provider.WorkerMessage)
	a.workerCalls.Store(message.ID, outputChan)
	defer a.workerCalls.Delete(message.ID)

	if err := provider.WriteMessage(ctx, a.mqttClient, a.publishTopic, message); err != nil {
		return provider.EmptyWorkerMessage, err
	}

	select {
	case output := <-outputChan:
		return output, nil
	case <-time.After(workerWaitDelay):
		return provider.EmptyWorkerMessage, ErrWorkerResponseTimeout
	}
}
