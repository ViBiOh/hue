package v2

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"go.opentelemetry.io/otel/metric"
)

type Service struct {
	lights        map[string]*Light
	groups        map[string]Group
	motionSensors map[string]MotionSensor
	taps          map[string]Tap

	temperatureMetric metric.Float64Gauge
	batteryMetric     metric.Int64Gauge
	motionMetric      metric.Int64Gauge
	lightLevelMetric  metric.Int64Gauge

	config homeConfig

	req   request.Request
	mutex sync.RWMutex
}

type Config struct {
	bridgeIP       string
	bridgeUsername string
	config         string
}

type homeConfig struct {
	Temperatures map[string]string
}

var errNoConfig = errors.New("no v2 config")

func Flags(fs *flag.FlagSet, prefix string) *Config {
	var config Config

	flags.New("BridgeIP", "IP of Bridge").Prefix(prefix).DocPrefix("hue").StringVar(fs, &config.bridgeIP, "", nil)
	flags.New("Username", "Username for Bridge").Prefix(prefix).DocPrefix("hue").StringVar(fs, &config.bridgeUsername, "", nil)
	flags.New("Config", "Configuration filename").Prefix(prefix).DocPrefix("hue").StringVar(fs, &config.config, "", nil)

	return &config
}

func New(config *Config, meterProvider metric.MeterProvider) (*Service, error) {
	service := &Service{
		req: request.Get(fmt.Sprintf("https://%s", config.bridgeIP)).Header("hue-application-key", config.bridgeUsername).WithClient(createInsecureClient(10 * time.Second)),
	}

	var err error

	if err := service.createMetrics(meterProvider); err != nil {
		return nil, fmt.Errorf("metric: %w", err)
	}

	service.config, err = loadConfig(config.config)
	if err != nil && !errors.Is(err, errNoConfig) {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return service, nil
}

func loadConfig(filename string) (homeConfig, error) {
	if len(filename) == 0 {
		return homeConfig{}, errNoConfig
	}

	configFile, err := os.Open(filename)
	if err != nil {
		return homeConfig{}, fmt.Errorf("open: %w", err)
	}

	var config homeConfig

	return config, json.NewDecoder(configFile).Decode(&config)
}
