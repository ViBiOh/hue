package v2

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"go.opentelemetry.io/otel/metric"
)

// App stores information and secret of API
type App struct {
	lights        map[string]*Light
	groups        map[string]Group
	motionSensors map[string]MotionSensor

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
		bridgeIP:       flags.New("BridgeIP", "IP of Bridge").Prefix(prefix).DocPrefix("hue").String(fs, "", nil),
		bridgeUsername: flags.New("Username", "Username for Bridge").Prefix(prefix).DocPrefix("hue").String(fs, "", nil),
		config:         flags.New("Config", "Configuration filename").Prefix(prefix).DocPrefix("hue").String(fs, "", nil),
	}
}

// New creates new App from Config
func New(config Config, meterProvider metric.MeterProvider) (*App, error) {
	bridgeAddress := strings.TrimSpace(*config.bridgeIP)
	bridgeUsername := strings.TrimSpace(*config.bridgeUsername)

	app := &App{
		req: request.Get(fmt.Sprintf("https://%s", bridgeAddress)).Header("hue-application-key", bridgeUsername).WithClient(createInsecureClient(10 * time.Second)),
	}

	err := app.createMetrics(meterProvider)
	if err != nil {
		return nil, err
	}

	return app, nil
}
