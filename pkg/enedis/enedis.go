package enedis

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// App of package
type App struct {
	consumption *Consumption
	mutex       sync.Mutex

	prometheus           bool
	prometheusCollectors map[string]prometheus.Gauge
}

// New creates new App from Config
func New() *App {
	return &App{}
}

// EnablePrometheus start prometheus register
func (a *App) EnablePrometheus() {
	a.prometheus = true
	a.prometheusCollectors = make(map[string]prometheus.Gauge)
}

// GetData returns data to be displayed
func (a *App) GetData() interface{} {
	return nil
}
