package hue

import (
	"flag"
	"fmt"
	"maps"
	"net/http"
	"sort"
	"sync"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	v2 "github.com/ViBiOh/hue/pkg/v2"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	v2Service      *v2.Service
	scenes         map[string]Scene
	schedules      map[string]Schedule
	renderer       *renderer.Service
	tracerProvider trace.TracerProvider
	bridgeUsername string
	bridgeURL      string
	configFileName string
	mutex          sync.RWMutex
	update         bool
}

type Config struct {
	BridgeIP       string
	BridgeUsername string
	Config         string
	Update         bool
}

func Flags(fs *flag.FlagSet, prefix string) *Config {
	var config Config

	flags.New("BridgeIP", "IP of Bridge").Prefix(prefix).DocPrefix("hue").StringVar(fs, &config.BridgeIP, "", nil)
	flags.New("Username", "Username for Bridge").Prefix(prefix).DocPrefix("hue").StringVar(fs, &config.BridgeUsername, "", nil)
	flags.New("Config", "Configuration filename").Prefix(prefix).DocPrefix("hue").StringVar(fs, &config.Config, "", nil)
	flags.New("Update", "Update configuration from file").Prefix(prefix).DocPrefix("hue").BoolVar(fs, &config.Update, false, nil)

	return &config
}

func New(config *Config, tracerProvider trace.TracerProvider, rendererService *renderer.Service, v2Service *v2.Service) (*Service, error) {
	service := Service{
		bridgeURL:      fmt.Sprintf("http://%s/api/%s", config.BridgeIP, config.BridgeUsername),
		bridgeUsername: config.BridgeUsername,
		configFileName: config.Config,
		update:         config.Update,
		renderer:       rendererService,
		tracerProvider: tracerProvider,
		v2Service:      v2Service,
	}

	return &service, nil
}

func (s *Service) TemplateFunc(_ http.ResponseWriter, _ *http.Request) (renderer.Page, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return renderer.NewPage("public", http.StatusOK, map[string]any{
		"Groups":    s.v2Service.Groups(),
		"Scenes":    s.toScenes(),
		"Schedules": s.toSchedules(),
		"Sensors":   s.v2Service.Sensors(),
	}), nil
}

func (s *Service) toScenes() map[string]Scene {
	output := make(map[string]Scene, len(s.scenes))

	maps.Copy(output, s.scenes)

	return output
}

func (s *Service) toSchedules() []Schedule {
	output := make([]Schedule, len(s.schedules))

	i := 0
	for _, item := range s.schedules {
		output[i] = item
		i++
	}

	sort.Sort(ByScheduleID(output))

	return output
}
