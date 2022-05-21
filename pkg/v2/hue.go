package v2

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/prometheus/client_golang/prometheus"
)

// App stores informations and secret of API
type App struct {
	groups        map[string]Group
	motionSensors map[string]MotionSensor
	metrics       map[string]*prometheus.GaugeVec

	req   request.Request
	mutex sync.RWMutex
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
		bridgeIP:       flags.String(fs, prefix, "hue", "BridgeIP", "IP of Bridge", "", nil),
		bridgeUsername: flags.String(fs, prefix, "hue", "Username", "Username for Bridge", "", nil),
		config:         flags.String(fs, prefix, "hue", "Config", "Configuration filename", "", nil),
	}
}

// New creates new App from Config
func New(config Config, prometheusRegisterer prometheus.Registerer) (*App, error) {
	metrics, err := createMetrics(prometheusRegisterer, "temperature", "battery")
	if err != nil {
		return nil, err
	}

	bridgeAddress := strings.TrimSpace(*config.bridgeIP)
	bridgeUsername := strings.TrimSpace(*config.bridgeUsername)

	app := App{
		req:     request.Get(fmt.Sprintf("https://%s", bridgeAddress)).Header("hue-application-key", bridgeUsername).WithClient(createInsecureClient(10 * time.Second)),
		metrics: metrics,
	}

	return &app, nil
}
