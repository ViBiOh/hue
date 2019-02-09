package dyson

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ViBiOh/iot/pkg/provider"
)

const (
	// Source constant for worker message
	Source = `dyson`
)

// App of package
type App struct {
	hub     provider.Hub
	devices []*Device
	mutex   sync.RWMutex
}

// New creates new App from Config
func New() *App {
	return &App{}
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.hub.RenderDashboard(w, r, http.StatusServiceUnavailable, &provider.Message{Level: `error`, Content: fmt.Sprintf(`[%s] Unknown command`, Source)})
	})
}

// SetHub receive Hub during init of it
func (a *App) SetHub(hub provider.Hub) {
	a.hub = hub
}

// GetWorkerSource get source of message
func (a *App) GetWorkerSource() string {
	return Source
}

// EnablePrometheus start prometheus register
func (a *App) EnablePrometheus() {
}

// GetData return data for Dashboard rendering
func (a *App) GetData() interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if len(a.devices) == 0 {
		return false
	}

	return &Data{
		Devices: a.devices,
	}
}
