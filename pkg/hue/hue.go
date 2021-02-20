package hue

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/prometheus/client_golang/prometheus"
)

// App stores informations and secret of API
type App interface {
	Handler() http.Handler
	TemplateFunc(*http.Request) (string, int, map[string]interface{}, error)
	Start(<-chan struct{})
}

// Config of package
type Config struct {
	bridgeIP       *string
	bridgeUsername *string
	config         *string
}

type app struct {
	config *configHue

	groups    map[string]Group
	scenes    map[string]Scene
	schedules map[string]Schedule
	sensors   map[string]Sensor

	renderer             renderer.App
	prometheusRegisterer prometheus.Registerer
	prometheusCollectors map[string]prometheus.Gauge

	bridgeURL      string
	bridgeUsername string

	mutex sync.RWMutex
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		bridgeIP:       flags.New(prefix, "hue").Name("BridgeIP").Default("").Label("IP of Bridge").ToString(fs),
		bridgeUsername: flags.New(prefix, "hue").Name("Username").Default("").Label("Username for Bridge").ToString(fs),
		config:         flags.New(prefix, "hue").Name("Config").Default("").Label("Configuration filename").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, registerer prometheus.Registerer, renderer renderer.App) (App, error) {
	bridgeUsername := strings.TrimSpace(*config.bridgeUsername)

	app := &app{
		bridgeURL:      fmt.Sprintf("http://%s/api/%s", strings.TrimSpace(*config.bridgeIP), bridgeUsername),
		bridgeUsername: bridgeUsername,

		renderer: renderer,

		prometheusRegisterer: registerer,
		prometheusCollectors: make(map[string]prometheus.Gauge),
	}

	configFile := strings.TrimSpace(*config.config)
	if len(configFile) != 0 {
		rawConfig, err := ioutil.ReadFile(configFile)
		if err != nil {
			return app, err
		}

		if err := json.Unmarshal(rawConfig, &app.config); err != nil {
			return app, err
		}
	}

	return app, nil
}

func (a *app) TemplateFunc(_ *http.Request) (string, int, map[string]interface{}, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return "public", http.StatusOK, map[string]interface{}{
		"Groups":    a.groups,
		"Scenes":    a.scenes,
		"Schedules": a.schedules,
		"Sensors":   a.sensors,
	}, nil
}
