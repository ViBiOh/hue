package hue

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/hue/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ App = &app{}
)

// App stores informations and secret of API
type App interface {
	Handle(*http.Request) (model.Message, int)
	Data() map[string]interface{}
	Start()
}

// Config of package
type Config struct {
	bridgeIP       *string
	bridgeUsername *string
	config         *string
}

type app struct {
	config *configHue
	cron   *cron.Cron

	groups    map[string]Group
	scenes    map[string]Scene
	schedules map[string]Schedule
	sensors   map[string]Sensor

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
func New(config Config, registerer prometheus.Registerer) (App, error) {
	bridgeUsername := strings.TrimSpace(*config.bridgeUsername)

	app := &app{
		bridgeURL:      fmt.Sprintf("http://%s/api/%s", strings.TrimSpace(*config.bridgeIP), bridgeUsername),
		bridgeUsername: bridgeUsername,

		prometheusRegisterer: registerer,
		prometheusCollectors: make(map[string]prometheus.Gauge),
	}

	if *config.config != "" {
		rawConfig, err := ioutil.ReadFile(*config.config)
		if err != nil {
			return app, err
		}

		if err := json.Unmarshal(rawConfig, &app.config); err != nil {
			return app, err
		}
	}

	return app, nil
}

func (a *app) Data() map[string]interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return map[string]interface{}{
		"Groups":    a.groups,
		"Scenes":    a.scenes,
		"Schedules": a.schedules,
		"Sensors":   a.sensors,
	}
}
