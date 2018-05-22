package iot

import (
	"context"
	"time"

	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	workerWaitDelay = 10 * time.Second
)

// SendToWorker sends payload to worker
func (a *App) SendToWorker(ctx context.Context, message *provider.WorkerMessage, waitOutput bool) *provider.WorkerMessage {
	var outputChan chan *provider.WorkerMessage

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
