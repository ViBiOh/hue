package netatmo

import (
	"encoding/json"
	"sync"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/iot/pkg/provider"
)

// App of package
type App struct {
	devices []*Device
	mutex   sync.RWMutex
}

// New creates new App from Config
func New() *App {
	return &App{
		devices: nil,
	}
}

// GetData return data for Dashboard rendering
func (a *App) GetData() interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.devices
}

// GetWorkerSource returns source for worker
func (a *App) GetWorkerSource() string {
	return Source
}

// WorkerHandler handler worker requests
func (a *App) WorkerHandler(p *provider.WorkerMessage) error {
	if p.Action == DevicesAction {
		return a.handleDevicesWorker(p)
	}

	return provider.ErrWorkerUnknownAction
}

func (a *App) handleDevicesWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var data []*Device
	if err := json.Unmarshal([]byte(message.Payload), &data); err != nil {
		return errors.WithStack(err)
	}

	a.devices = data

	return nil
}
