package iot

import (
	"context"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	workerWaitDelay = 10 * time.Second
)

func (a *App) registerWorker(worker provider.WorkerProvider) {
	a.workerProviders[worker.GetWorkerSource()] = worker
}

// SendToWorker sends payload to worker
func (a *App) SendToWorker(ctx context.Context, source, messageType string, payload interface{}, waitOutput bool) *provider.WorkerMessage {
	var outputChan chan *provider.WorkerMessage

	message := &provider.WorkerMessage{
		ID:      tools.Sha1(payload),
		Source:  source,
		Type:    messageType,
		Payload: fmt.Sprintf(`%s`, payload),
	}

	if waitOutput {
		outputChan = make(chan *provider.WorkerMessage)
		a.workerCalls.Store(message.ID, outputChan)

		defer a.workerCalls.Delete(message.ID)
	}

	if err := provider.WriteMessage(ctx, a.wsConn, message); err != nil {
		return &provider.WorkerMessage{
			Source:  message.Source,
			Type:    provider.WorkerErrorType,
			Payload: err,
		}
	}

	if waitOutput {
		select {
		case output := <-outputChan:
			return output
		case <-time.After(workerWaitDelay):
			return nil
		}
	}

	return nil
}
