package dyson

import (
	"encoding/json"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	// WorkerDevicesAction message prefix for groups command
	WorkerDevicesAction = "devices"
)

func (a *App) handleDevicesFromWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var newDevices []*Device
	if err := json.Unmarshal([]byte(message.Payload), &newDevices); err != nil {
		return errors.WithStack(err)
	}

	a.devices = newDevices
	if a.prometheus {
		a.updatePrometheus()
	}

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(p *provider.WorkerMessage) error {
	if p.Action == WorkerDevicesAction {
		return a.handleDevicesFromWorker(p)
	}

	return provider.ErrWorkerUnknownAction
}
