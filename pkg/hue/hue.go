package hue

import (
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/prometheus/client_golang/prometheus"
)

// App stores informations and secret of API
type App struct {
	apiHandler http.Handler
	metrics    map[string]*prometheus.GaugeVec

	syncers []syncer

	lights    map[string]Light
	groups    map[string]Group
	scenes    map[string]Scene
	schedules map[string]Schedule
	sensors   map[string]Sensor

	bridgeUsername string
	bridgeURL      string
	configFileName string

	rendererApp renderer.App
	mutex       sync.RWMutex
}

// Config of package
type Config struct {
	bridgeIP       *string
	bridgeUsername *string
	config         *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		bridgeIP:       flags.New(prefix, "hue", "BridgeIP").Default("", nil).Label("IP of Bridge").ToString(fs),
		bridgeUsername: flags.New(prefix, "hue", "Username").Default("", nil).Label("Username for Bridge").ToString(fs),
		config:         flags.New(prefix, "hue", "Config").Default("", nil).Label("Configuration filename").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, prometheusRegisterer prometheus.Registerer, renderer renderer.App) (*App, error) {
	metrics, err := createMetrics(prometheusRegisterer, "temperature", "battery")
	if err != nil {
		return nil, err
	}

	bridgeUsername := strings.TrimSpace(*config.bridgeUsername)

	app := App{
		bridgeURL:      fmt.Sprintf("http://%s/api/%s", strings.TrimSpace(*config.bridgeIP), bridgeUsername),
		bridgeUsername: bridgeUsername,
		configFileName: strings.TrimSpace(*config.config),
		rendererApp:    renderer,
		metrics:        metrics,
	}

	app.syncers = []syncer{
		app.syncGroups,
		app.syncSchedules,
		app.syncSensors,
		app.syncScenes,
	}

	app.apiHandler = http.StripPrefix(apiPath, app.Handler())

	return &app, nil
}

// TemplateFunc for rendering GUI
func (a *App) TemplateFunc(w http.ResponseWriter, r *http.Request) (string, int, map[string]interface{}, error) {
	if strings.HasPrefix(r.URL.Path, apiPath) {
		a.apiHandler.ServeHTTP(w, r)
		return "", 0, nil, nil
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return "public", http.StatusOK, map[string]interface{}{
		"Groups":    a.toGroups(),
		"Scenes":    a.toScenes(),
		"Schedules": a.toSchedules(),
		"Sensors":   a.toSensors(),
	}, nil
}

func (a *App) toGroups() map[string]Group {
	output := make(map[string]Group, len(a.groups))

	for key, item := range a.groups {
		output[key] = item
	}

	return output
}

func (a *App) toScenes() map[string]Scene {
	output := make(map[string]Scene, len(a.scenes))

	for key, item := range a.scenes {
		output[key] = item
	}

	return output
}

func (a *App) toSchedules() []Schedule {
	output := make([]Schedule, len(a.schedules))

	i := 0
	for _, item := range a.schedules {
		output[i] = item
		i++
	}

	sort.Sort(ByScheduleID(output))

	return output
}

func (a *App) toSensors() []Sensor {
	output := make([]Sensor, len(a.sensors))

	i := 0
	for _, item := range a.sensors {
		output[i] = item
		i++
	}

	sort.Sort(BySensorID(output))

	return output
}
