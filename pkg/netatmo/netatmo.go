package netatmo

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ViBiOh/iot/pkg/provider"
)

// App stores informations and secret of API
type App struct {
	devices []*Device
	mutex   sync.RWMutex
}

// NewApp create Client from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		devices: nil,
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return nil
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

	convert, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf(`error while converting devices payload: %v`, err)
	}

	if err := json.Unmarshal(convert, &data); err != nil {
		return fmt.Errorf(`error while unmarshalling devices: %v`, err)
	}

	a.devices = data

	return nil
}
