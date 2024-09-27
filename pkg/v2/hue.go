package v2

import (
	"flag"
	"fmt"
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

	req   request.Request
	mutex sync.RWMutex
}

type Config struct {
	bridgeIP       string
	bridgeUsername string
	config         string
}

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

	if err := service.createMetrics(meterProvider); err != nil {
		return nil, fmt.Errorf("metric: %w", err)
	}

	return service, nil
}
